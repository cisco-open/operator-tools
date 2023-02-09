// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package reconciler

import (
	"context"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/cisco-open/operator-tools/pkg/resources"
	"github.com/cisco-open/operator-tools/pkg/types"
	"github.com/cisco-open/operator-tools/pkg/utils"
)

type ComponentReconciler interface {
	Reconcile(object runtime.Object) (*reconcile.Result, error)
	RegisterWatches(*builder.Builder)
}

type Watches interface {
	SetupAdditionalWatches(c controller.Controller) error
}

type ComponentWithStatus interface {
	Update(object runtime.Object, status types.ReconcileStatus, msg string) error
	IsSkipped(object runtime.Object) bool
	IsEnabled(object runtime.Object) bool
}

type ComponentLifecycle interface {
	OnFinished(object runtime.Object) error
}

// ComponentReconcilers is a list of component reconcilers that support getting components in Install and Uninstall order.
type ComponentReconcilers []ComponentReconciler

func (c ComponentReconcilers) Get(order utils.ResourceOrder) []ComponentReconciler {
	if order != utils.UninstallResourceOrder {
		return c
	}

	components := []ComponentReconciler{}
	for i := len(c) - 1; i >= 0; i-- {
		components = append(components, c[i])
	}
	return components
}

// Dispatcher orchestrates reconciliation of multiple ComponentReconciler objects
// focusing on handing off reconciled object to all of its components and calculating an aggregated result to return.
// It requires a ResourceGetter callback and optionally can leverage a ResourceFilter and a CompletionHandler
type Dispatcher struct {
	client.Client
	Log                  logr.Logger
	ResourceGetter       func(req ctrl.Request) (runtime.Object, error)
	ResourceFilter       func(runtime.Object) (bool, error)
	CompletionHandler    func(runtime.Object, ctrl.Result, error) (ctrl.Result, error)
	ComponentReconcilers ComponentReconcilers
	// ForceResourceOrder can be used to force a given resource ordering regardless of an object being deleted with
	// finalizers.
	ForceResourceOrder utils.ResourceOrder
}

// Reconcile implements reconcile.Reconciler in a generic way from the controller-runtime library
func (r *Dispatcher) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	object, err := r.ResourceGetter(req)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}
	if object == nil {
		return reconcile.Result{}, nil
	}
	if r.ResourceFilter != nil {
		shouldReconcile, err := r.ResourceFilter(object)
		if err != nil || !shouldReconcile {
			return reconcile.Result{}, errors.WithStack(err)
		}
	}
	result, err := r.Handle(object)
	if r.CompletionHandler != nil {
		return r.CompletionHandler(object, result, errors.WithStack(err))
	}
	if err != nil {
		return result, errors.WithStack(err)
	}
	return result, nil
}

// Handle receives a single object and dispatches it to all the components
// Components need to understand how to interpret the object
func (r *Dispatcher) Handle(object runtime.Object) (ctrl.Result, error) {
	isBeingDeleted, err := resources.IsObjectBeingDeleted(object)
	if err != nil {
		return ctrl.Result{}, err
	}

	componentExecutionOrder := utils.InstallResourceOrder
	if isBeingDeleted {
		componentExecutionOrder = utils.UninstallResourceOrder
	}

	if r.ForceResourceOrder != "" {
		componentExecutionOrder = r.ForceResourceOrder
	}

	combinedResult := &CombinedResult{}
	for _, cr := range r.ComponentReconcilers.Get(componentExecutionOrder) {
		if cr, ok := cr.(ComponentWithStatus); ok {
			status := types.ReconcileStatusReconciling
			if cr.IsSkipped(object) {
				status = types.ReconcileStatusUnmanaged
			}
			if uerr := cr.Update(object, status, ""); uerr != nil {
				combinedResult.CombineErr(errors.WrapIf(uerr, "unable to update status for component"))
			}
			if cr.IsSkipped(object) {
				continue
			}
		}

		// Any patch/update command can update the object, thus sequent steps will be
		// remove steps instead. To work around that let's requeue with the correct order.
		currentDeletionStatus, err := resources.IsObjectBeingDeleted(object)
		if err != nil {
			return ctrl.Result{}, err
		}

		if currentDeletionStatus != isBeingDeleted {
			// Requeue object so that we can start reconciling in inverse order
			combinedResult.Combine(&reconcile.Result{Requeue: true, RequeueAfter: 0}, errors.New("object being deleted requeing object"))
			break
		}

		result, err := cr.Reconcile(object)
		if cr, ok := cr.(ComponentWithStatus); ok {
			if err != nil {
				if uerr := cr.Update(object, types.ReconcileStatusFailed, err.Error()); uerr != nil {
					combinedResult.CombineErr(errors.WrapIf(uerr, "unable to update status for component"))
				}
			} else {
				if result == nil || (!result.Requeue && result.RequeueAfter == 0) {
					status := types.ReconcileStatusRemoved
					if cr.IsEnabled(object) {
						status = types.ReconcileStatusAvailable
					}
					if uerr := cr.Update(object, status, ""); uerr != nil {
						combinedResult.CombineErr(errors.WrapIf(uerr, "unable to update status for component"))
					}
				}
			}
		}
		if cr, ok := cr.(ComponentLifecycle); ok {
			if err := cr.OnFinished(object); err != nil {
				combinedResult.Combine(result, errors.WrapIf(err, "failed to notify component on finish"))
			}
		}
		combinedResult.Combine(result, errors.WithStack(err))
		if cr, ok := cr.(interface{ IsOptional() bool }); ok {
			if err != nil && !cr.IsOptional() {
				break
			}
		}
	}
	return combinedResult.Result, combinedResult.Err
}

// RegisterWatches dispatches the watch registration builder to all its components
func (r *Dispatcher) RegisterWatches(b *builder.Builder) *builder.Builder {
	for _, cr := range r.ComponentReconcilers {
		cr.RegisterWatches(b)
	}
	return b
}

// SetupAdditionalWatches dispatches the controller for watch registration to all its components
func (r *Dispatcher) SetupAdditionalWatches(c controller.Controller) error {
	for _, cr := range r.ComponentReconcilers {
		if cr, ok := cr.(Watches); ok {
			err := cr.SetupAdditionalWatches(c)
			if err != nil {
				return errors.WithStack(err)
			}
		}
	}

	return nil
}

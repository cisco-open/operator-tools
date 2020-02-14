// Copyright © 2019 Banzai Cloud
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
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ComponentReconciler interface {
	Reconcile(object runtime.Object) (*reconcile.Result, error)
	RegisterWatches(*builder.Builder)
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
	ComponentReconcilers []ComponentReconciler
}

// Reconcile implements reconcile.Reconciler in a generic way from the controller-runtime library
func (r *Dispatcher) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	object, err := r.ResourceGetter(req)
	if err != nil {
		return reconcile.Result{}, err
	}
	if r.ResourceFilter != nil {
		shouldReconcile, err := r.ResourceFilter(object)
		if err != nil || !shouldReconcile {
			return reconcile.Result{}, err
		}
	}
	result, err := r.Handle(object)
	if r.CompletionHandler != nil {
		return r.CompletionHandler(object, result, err)
	}
	if err != nil {
		return result, err
	}
	return result, nil
}

// Handle receives a single object and dispatches it to all the components
// Components need to understand how to interpret the object
func (r *Dispatcher) Handle(object runtime.Object) (ctrl.Result, error) {
	combinedResult := &CombinedResult{}
	for _, cr := range r.ComponentReconcilers {
		result, err := cr.Reconcile(object)
		combinedResult.Combine(result, err)
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

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
	"emperror.dev/errors"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/banzaicloud/operator-tools/pkg/resources"
	"github.com/banzaicloud/operator-tools/pkg/wait"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

type Reconciler struct {
	mgr     manager.Manager
	log     logr.Logger
	baseLog logr.Logger

	waitForResourcesEnabled bool
	waitBackoff             wait.Backoff
}

type ReconcilerOption func(*Reconciler)

func NewReconciler(
	mgr manager.Manager,
	log logr.Logger, options ...ReconcilerOption) *Reconciler {

	r := &Reconciler{
		mgr:     mgr,
		log:     log,
		baseLog: log,
	}

	for _, opt := range options {
		opt(r)
	}

	return r
}

func WaitForResources(backoff wait.Backoff) ReconcilerOption {
	return func(r *Reconciler) {
		r.waitForResourcesEnabled = true
		r.waitBackoff = backoff
	}
}

func (rec *Reconciler) IsWaitForResourcesEnabled() bool {
	return rec.waitForResourcesEnabled
}

func (rec *Reconciler) Manager() manager.Manager {
	return rec.mgr
}

func (rec *Reconciler) SetManager(mgr manager.Manager) {
	rec.mgr = mgr
}

func (rec *Reconciler) Logger() logr.Logger {
	return rec.log
}

func (rec *Reconciler) BaseLogger() logr.Logger {
	return rec.baseLog
}

func (rec *Reconciler) SetLogger(log logr.Logger) {
	rec.log = log
}

func (r *Reconciler) RegisterWatches(b *builder.Builder) {}

func (rec *Reconciler) GetResourceBuildersFromObjects(objects []runtime.Object, state DesiredState, modifierFuncs ...resources.ObjectModifierFunc) ([]ResourceBuilder, error) {
	resources := []ResourceBuilder{}

	utils.RuntimeObjects(objects).Sort(utils.InstallResourceOrder)

	for _, o := range objects {
		o := o
		for _, modifierFunc := range modifierFuncs {
			var err error
			o, err = modifierFunc(o)
			if err != nil {
				return nil, err
			}
		}
		resources = append(resources, func() (runtime.Object, DesiredState, error) {
			if desired, ok := o.(*corev1.Service); ok {
				beforeUpdateHook := DesiredStateHook(func(current runtime.Object) error {
					if s, ok := current.(*corev1.Service); ok {
						desired.Spec.ClusterIP = s.Spec.ClusterIP
					} else {
						return errors.Errorf("failed to cast service object %+v", current)
					}
					return nil
				})
				return o, beforeUpdateHook, nil
			}

			return o, state, nil
		})
	}

	return resources, nil
}

func (rec *Reconciler) GetNativeReconciler(
	component string,
	resourceBuilders ResourceBuilders,
	purgeTypes PurgeTypesFunc,
	resourceTranslate ResourceTranslate,
	reconcilerOpts *ReconcilerOpts,
) *NativeReconciler {
	if reconcilerOpts == nil {
		reconcilerOpts = &ReconcilerOpts{
			EnableRecreateWorkloadOnImmutableFieldChange: true,
			Scheme: rec.mgr.GetScheme(),
		}
	}

	return NewNativeReconciler(
		component,
		NewGenericReconciler(
			rec.mgr.GetClient(),
			rec.log,
			*reconcilerOpts,
		),
		rec.mgr.GetClient(),
		NewReconciledComponent(
			resourceBuilders,
			nil,
			purgeTypes,
		),
		resourceTranslate,
		NativeReconcilerWithScheme(rec.mgr.GetScheme()),
	)
}

func (rec *Reconciler) WaitForResources(r *NativeReconciler, presentObjects, absentObjects []runtime.Object) error {
	if !rec.waitForResourcesEnabled {
		return nil
	}

	rcc := wait.NewResourceConditionChecks(rec.mgr.GetClient(), rec.waitBackoff, rec.log.WithName("wait"), rec.mgr.GetScheme())

	err := rcc.WaitForResources("readiness", presentObjects, wait.ExistsConditionCheck, wait.ReadyReplicasConditionCheck)
	if err != nil {
		return err
	}

	absentObjects = append(absentObjects, r.GetReconciledObjectWithState(ReconciledObjectStatePurged)...)
	err = rcc.WaitForResources("removal", absentObjects, wait.NonExistsConditionCheck)
	if err != nil {
		return err
	}

	return nil
}

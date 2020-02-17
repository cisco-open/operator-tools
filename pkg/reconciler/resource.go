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
	"context"
	"fmt"
	"reflect"
	"time"

	"emperror.dev/errors"
	"github.com/banzaicloud/k8s-objectmatcher/patch"
	"github.com/go-logr/logr"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	StateAbsent  StaticDesiredState = "Absent"
	StatePresent StaticDesiredState = "Present"
)

type DesiredState interface {
	BeforeUpdate(object runtime.Object) error
}

type ResourceReconciler interface {
	CreateIfNotExist(runtime.Object) (created bool, object runtime.Object, err error)
	ReconcileResource(runtime.Object, DesiredState) (*reconcile.Result, error)
}

type StaticDesiredState string

func (s StaticDesiredState) BeforeUpdate(object runtime.Object) error {
	return nil
}

type DesiredStateHook func(object runtime.Object) error

func (d DesiredStateHook) BeforeUpdate(object runtime.Object) error {
	return d(object)
}

// GenericResourceReconciler generic resource reconciler
type GenericResourceReconciler struct {
	Log     logr.Logger
	Client  runtimeClient.Client
	Options ReconcilerOpts
}

type ResourceReconcilerOption func(*ReconcilerOpts)

// Recommended to use NewReconcilerWith + ResourceReconcilerOptions
type ReconcilerOpts struct {
	Log                                              logr.Logger
	Scheme                                           *runtime.Scheme
	EnableRecreateWorkloadOnImmutableFieldChange     bool
	EnableRecreateWorkloadOnImmutableFieldChangeHelp string
}

// NewReconciler returns GenericResourceReconciler
func NewReconciler(client runtimeClient.Client, log logr.Logger, opts ReconcilerOpts) *GenericResourceReconciler {
	if opts.Scheme == nil {
		opts.Scheme = runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(opts.Scheme)
	}
	return &GenericResourceReconciler{
		Log:     log,
		Client:  client,
		Options: opts,
	}
}

func WithLog(log logr.Logger) ResourceReconcilerOption {
	return func(o *ReconcilerOpts) {
		o.Log = log
	}
}

func WithScheme(scheme *runtime.Scheme) ResourceReconcilerOption {
	return func(o *ReconcilerOpts) {
		o.Scheme = scheme
	}
}

func WithEnableRecreateWorkload() ResourceReconcilerOption {
	return func(o *ReconcilerOpts) {
		o.EnableRecreateWorkloadOnImmutableFieldChange = true
	}
}

func NewReconcilerWith(client runtimeClient.Client, opts ...func(reconciler *ReconcilerOpts)) ResourceReconciler {
	rec := &GenericResourceReconciler{
		Client: client,
		Options: ReconcilerOpts{
			EnableRecreateWorkloadOnImmutableFieldChangeHelp: "recreating object on immutable field change has to be enabled explicitly through the reconciler options",
		},
		Log: log.NullLogger{},
	}
	for _, opt := range opts {
		opt(&rec.Options)
	}
	if rec.Options.Log != nil {
		rec.Log = rec.Options.Log
	}
	if rec.Options.Scheme == nil {
		rec.Options.Scheme = runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(rec.Options.Scheme)
	}
	return rec
}

// CreateResource creates a resource if it doesn't exist
func (r *GenericResourceReconciler) CreateResource(desired runtime.Object) error {
	_, _, err := r.CreateIfNotExist(desired)
	return err
}

// ReconcileResource reconciles various kubernetes types
func (r *GenericResourceReconciler) ReconcileResource(desired runtime.Object, desiredState DesiredState) (*reconcile.Result, error) {
	resourceDetails, err := r.resourceDetails(desired)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get resource details")
	}
	log := r.resourceLog(desired, resourceDetails...)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to prepare resource logger")
	}
	debugLog := log.V(1)
	traceLog := log.V(2)
	switch desiredState {
	default:
		created, current, err := r.CreateIfNotExist(desired)
		if err == nil && created {
			return nil, nil
		}
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "failed to create resource", resourceDetails...)
		}

		// last chance to hook into the desired state armed with the knowledge of the current state
		err = desiredState.BeforeUpdate(current)
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "failed to get desired state dynamically", resourceDetails...)
		}
		if err == nil {
			if metaObject, ok := current.(metav1.Object); ok {
				if metaObject.GetDeletionTimestamp() != nil {
					log.Info(fmt.Sprintf("object %s is being deleted, backing off", metaObject.GetSelfLink()))
					return &reconcile.Result{RequeueAfter: time.Second * 2}, nil
				}
			}
			patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired)
			if err != nil {
				log.Error(err, "could not match objects")
			} else if patchResult.IsEmpty() {
				debugLog.Info("resource is in sync")
				return nil, nil
			} else {
				traceLog.Info("resource diffs",
					"patch", string(patchResult.Patch),
					"current", string(patchResult.Current),
					"modified", string(patchResult.Modified),
					"original", string(patchResult.Original))
			}

			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
				log.Error(err, "Failed to set last applied annotation", "desired", desired)
			}

			metaAccessor := meta.NewAccessor()

			currentResourceVersion, err := metaAccessor.ResourceVersion(current)
			if err != nil {
				return nil, errors.WrapIfWithDetails(err, "failed to access resourceVersion from metadata", resourceDetails...)
			}
			if err := metaAccessor.SetResourceVersion(desired, currentResourceVersion); err != nil {
				return nil, errors.WrapIfWithDetails(err, "failed to set resourceVersion in metadata", resourceDetails...)
			}

			debugLog.Info("Updating resource")
			if err := r.Client.Update(context.TODO(), desired); err != nil {
				sErr, ok := err.(*apierrors.StatusError)
				if ok && sErr.ErrStatus.Code == 422 && sErr.ErrStatus.Reason == metav1.StatusReasonInvalid {
					if r.Options.EnableRecreateWorkloadOnImmutableFieldChange {
						log.Error(err, "failed to update resource, trying to recreate")
						err := r.Client.Delete(context.TODO(), current,
							// wait until all dependent resources gets cleared up
							runtimeClient.PropagationPolicy(metav1.DeletePropagationForeground),
						)
						if err != nil {
							return nil, errors.WrapIf(err, "failed to delete current resource")
						}
						return &reconcile.Result{
							Requeue:      true,
							RequeueAfter: time.Second * 10,
						}, nil
					} else {
						return nil, errors.WrapIf(sErr, r.Options.EnableRecreateWorkloadOnImmutableFieldChangeHelp)
					}
				}
				return nil, errors.WrapIfWithDetails(err, "updating resource failed", resourceDetails...)
			}
			debugLog.Info("resource updated")
		}
	case StateAbsent:
		_, err := r.delete(desired)
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "failed to delete resource", resourceDetails...)
		}
	}
	return nil, nil
}

func (r *GenericResourceReconciler) CreateIfNotExist(desired runtime.Object) (bool, runtime.Object, error) {
	current := reflect.New(reflect.Indirect(reflect.ValueOf(desired)).Type()).Interface().(runtime.Object)
	key, err := runtimeClient.ObjectKeyFromObject(desired)
	if err != nil {
		return false, nil, errors.WrapIf(err, "failed to get object key")
	}
	resourceDetails, err := r.resourceDetails(desired)
	if err != nil {
		return false, nil, errors.WrapIf(err, "failed to get resource details")
	}
	log := r.resourceLog(desired, resourceDetails...)
	traceLog := log.V(2)
	err = r.Client.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return false, nil, errors.WrapIfWithDetails(err, "getting resource failed", resourceDetails...)
	}
	if apierrors.IsNotFound(err) {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
			log.Error(err, "Failed to set last applied annotation", "desired", desired)
		}
		if err := r.Client.Create(context.TODO(), desired); err != nil {
			return false, nil, errors.WrapIfWithDetails(err, "creating resource failed", resourceDetails...)
		}
		switch t := desired.DeepCopyObject().(type) {
		case *v1beta1.CustomResourceDefinition:
			err = wait.Poll(time.Second*1, time.Second*10, func() (done bool, err error) {
				err = r.Client.Get(context.TODO(), runtimeClient.ObjectKey{Namespace: t.Namespace, Name: t.Name}, t)
				if err != nil {
					return false, err
				}
				return crdReady(t), nil
			})
			if err != nil {
				return false, nil, errors.WrapIfWithDetails(err, "failed to wait for the crd to get ready", resourceDetails...)
			}
		}
		log.Info("resource created")
		return true, current, nil
	}
	traceLog.Info("resource already exists")
	return false, current, nil
}

func (r *GenericResourceReconciler) delete(desired runtime.Object) (bool, error) {
	key, err := runtimeClient.ObjectKeyFromObject(desired)
	if err != nil {
		return false, errors.WrapIf(err, "failed to get object key")
	}
	resourceDetails, err := r.resourceDetails(desired)
	if err != nil {
		return false, errors.WrapIf(err, "failed to get resource details")
	}
	log := r.resourceLog(desired, resourceDetails...)
	debugLog := log.V(1)
	traceLog := log.V(2)
	current := reflect.New(reflect.Indirect(reflect.ValueOf(desired)).Type()).Interface().(runtime.Object)
	err = r.Client.Get(context.TODO(), key, current)
	if err != nil {
		// If the resource type does not exist we should be ok to move on
		if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
			return false, nil
		}
		if !apierrors.IsNotFound(err) {
			return false, errors.WrapIfWithDetails(err, "getting resource failed", resourceDetails...)
		} else {
			traceLog.Info("resource not found skipping delete")
			return false, nil
		}
	}
	err = r.Client.Delete(context.TODO(), current)
	if err != nil {
		return false, errors.WrapIfWithDetails(err, "failed to delete resource", resourceDetails...)
	}
	debugLog.Info("resource deleted")
	return true, nil
}

func crdReady(crd *v1beta1.CustomResourceDefinition) bool {
	for _, cond := range crd.Status.Conditions {
		switch cond.Type {
		case v1beta1.Established:
			if cond.Status == v1beta1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

func (r *GenericResourceReconciler) resourceDetails(desired runtime.Object) ([]interface{}, error) {
	key, err := runtimeClient.ObjectKeyFromObject(desired)
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	values := []interface{}{"name", key.Name}
	if key.Namespace != "" {
		values = append(values, "namespace", key.Namespace)
	}
	defaultValues := append(values, "type", reflect.TypeOf(desired).String())
	if r.Options.Scheme == nil {
		return defaultValues, nil
	}
	var versionKinds []schema.GroupVersionKind
	versionKinds, _, err = r.Options.Scheme.ObjectKinds(desired)
	if len(versionKinds) == 0 || err != nil {
		r.Log.Error(err, "failed to get gvk for resource, falling back to type")
		return defaultValues, nil
	}
	if len(versionKinds) > 0 {
		values = append(values,
			"group", versionKinds[0].Group,
			"version", versionKinds[0].Version,
			"kind", versionKinds[0].Kind)
	}
	return values, nil
}

func (r *GenericResourceReconciler) resourceLog(desired runtime.Object, details ...interface{}) logr.Logger {
	if len(details) > 0 {
		return r.Log.WithValues(details...)
	}
	return r.Log
}

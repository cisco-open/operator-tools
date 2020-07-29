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
	"github.com/go-logr/logr"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
)

const (
	StateCreated StaticDesiredState = "Created"
	StateAbsent  StaticDesiredState = "Absent"
	StatePresent StaticDesiredState = "Present"
)

type DesiredState interface {
	BeforeUpdate(current, desired runtime.Object) error
	BeforeCreate(desired runtime.Object) error
	BeforeDelete(current runtime.Object) error
}

type DesiredStateShouldCreate interface {
	ShouldCreate(desired runtime.Object) (bool, error)
}

type DesiredStateShouldUpdate interface {
	ShouldUpdate(current, desired runtime.Object) (bool, error)
}

type DesiredStateShouldDelete interface {
	ShouldDelete(desired runtime.Object) (bool, error)
}

type DesiredStateWithDeleteOptions interface {
	GetDeleteOptions() []runtimeClient.DeleteOption
}

type DesiredStateWithCreateOptions interface {
	GetCreateOptions() []runtimeClient.CreateOption
}

type DesiredStateWithUpdateOptions interface {
	GetUpdateOptions() []runtimeClient.UpdateOption
}

type DesiredStateWithStaticState interface {
	DesiredState() StaticDesiredState
}

type ResourceReconciler interface {
	CreateIfNotExist(runtime.Object, DesiredState) (created bool, object runtime.Object, err error)
	ReconcileResource(runtime.Object, DesiredState) (*reconcile.Result, error)
}

type StaticDesiredState string

func (s StaticDesiredState) BeforeUpdate(current, desired runtime.Object) error {
	return nil
}

func (s StaticDesiredState) BeforeCreate(desired runtime.Object) error {
	return nil
}

func (s StaticDesiredState) BeforeDelete(current runtime.Object) error {
	return nil
}

type DesiredStateHook func(object runtime.Object) error

func (d DesiredStateHook) BeforeUpdate(current, desired runtime.Object) error {
	return d(current)
}

func (d DesiredStateHook) BeforeCreate(desired runtime.Object) error {
	return d(desired)
}

func (d DesiredStateHook) BeforeDelete(current runtime.Object) error {
	return d(current)
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
	_, _, err := r.CreateIfNotExist(desired, nil)
	return err
}

// ReconcileResource reconciles various kubernetes types
func (r *GenericResourceReconciler) ReconcileResource(desired runtime.Object, desiredState DesiredState) (*reconcile.Result, error) {
	resourceDetails, err := r.resourceDetails(desired)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to get resource details")
	}
	log := r.resourceLog(desired, resourceDetails...)
	debugLog := log.V(1)
	traceLog := log.V(2)
	state := desiredState
	if ds, ok := desiredState.(DesiredStateWithStaticState); ok {
		state = ds.DesiredState()
	}
	switch state {
	case StateCreated:
		created, _, err := r.CreateIfNotExist(desired, desiredState)
		if err == nil && created {
			return nil, nil
		}
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "failed to create resource", resourceDetails...)
		}
	default:
		created, current, err := r.CreateIfNotExist(desired, desiredState)
		if err == nil && created {
			return nil, nil
		}
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "failed to create resource", resourceDetails...)
		}

		if metaObject, ok := current.(metav1.Object); ok {
			if metaObject.GetDeletionTimestamp() != nil {
				log.Info(fmt.Sprintf("object %s is being deleted, backing off", metaObject.GetSelfLink()))
				return &reconcile.Result{RequeueAfter: time.Second * 2}, nil
			}
		}

		if ds, ok := desiredState.(DesiredStateShouldUpdate); ok {
			should, err := ds.ShouldUpdate(current.DeepCopyObject(), desired.DeepCopyObject())
			if err != nil {
				return nil, err
			}
			if !should {
				return nil, nil
			}
		}

		// last chance to hook into the desired state armed with the knowledge of the current state
		err = desiredState.BeforeUpdate(current, desired)
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "failed to get desired state dynamically", resourceDetails...)
		}

		patchResult, err := patch.DefaultPatchMaker.Calculate(current, desired, patch.IgnoreStatusFields())
		if err != nil {
			log.Error(err, "could not match objects")
		} else if patchResult.IsEmpty() {
			debugLog.Info("resource is in sync")
			return nil, nil
		} else {
			debugLog.Info("resource diff", "patch", string(patchResult.Patch))
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
		updateOptions := make([]runtimeClient.UpdateOption, 0)
		if ds, ok := desiredState.(DesiredStateWithUpdateOptions); ok {
			updateOptions = append(updateOptions, ds.GetUpdateOptions()...)
		}
		if err := r.Client.Update(context.TODO(), desired, updateOptions...); err != nil {
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

	case StateAbsent:
		_, err := r.delete(desired, desiredState)
		if err != nil {
			return nil, errors.WrapIfWithDetails(err, "failed to delete resource", resourceDetails...)
		}
	}
	return nil, nil
}

func (r *GenericResourceReconciler) fromDesired(desired runtime.Object) (runtime.Object, error) {
	if _, ok := desired.(*unstructured.Unstructured); ok {
		if r.Options.Scheme != nil {
			object, err := r.Options.Scheme.New(desired.GetObjectKind().GroupVersionKind())
			if err == nil {
				return object, nil
			}
			r.Log.V(2).Info("unable to detect correct type for the resource, falling back to unstructured")
		}
		current := &unstructured.Unstructured{}
		desiredGVK := desired.GetObjectKind()
		current.SetKind(desiredGVK.GroupVersionKind().Kind)
		current.SetAPIVersion(desiredGVK.GroupVersionKind().GroupVersion().String())
		return current, nil
	}
	return reflect.New(reflect.Indirect(reflect.ValueOf(desired)).Type()).Interface().(runtime.Object), nil
}

func (r *GenericResourceReconciler) CreateIfNotExist(desired runtime.Object, desiredState DesiredState) (bool, runtime.Object, error) {
	current, err := r.fromDesired(desired)
	if err != nil {
		return false, nil, errors.WrapIf(err, "failed to create new object based on desired")
	}
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
		if desiredState != nil {
			err = desiredState.BeforeCreate(desired)
			if err != nil {
				return false, nil, errors.WrapIfWithDetails(err, "failed to prepare desired state before creation", resourceDetails...)
			}
			if ds, ok := desiredState.(DesiredStateShouldCreate); ok {
				should, err := ds.ShouldCreate(desired)
				if err != nil {
					return false, desired, err
				}
				if !should {
					return false, desired, nil
				}
			}
		}
		createOptions := make([]runtimeClient.CreateOption, 0)
		if ds, ok := desiredState.(DesiredStateWithCreateOptions); ok {
			createOptions = append(createOptions, ds.GetCreateOptions()...)
		}
		if err := r.Client.Create(context.TODO(), desired, createOptions...); err != nil {
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

func (r *GenericResourceReconciler) delete(desired runtime.Object, desiredState DesiredState) (bool, error) {
	current, err := r.fromDesired(desired)
	if err != nil {
		return false, errors.WrapIf(err, "failed to create new object based on desired")
	}
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
	if desiredState != nil {
		err = desiredState.BeforeDelete(current)
		if err != nil {
			return false, errors.WrapIfWithDetails(err, "failed to prepare desired state before deletion", resourceDetails...)
		}
		if ds, ok := desiredState.(DesiredStateShouldDelete); ok {
			should, err := ds.ShouldDelete(desired)
			if err != nil {
				return false, err
			}
			if !should {
				return false, nil
			}
		}
	}
	deleteOptions := make([]runtimeClient.DeleteOption, 0)
	if ds, ok := desiredState.(DesiredStateWithDeleteOptions); ok {
		deleteOptions = append(deleteOptions, ds.GetDeleteOptions()...)
	}
	err = r.Client.Delete(context.TODO(), current, deleteOptions...)
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
	gvk, err := apiutil.GVKForObject(desired, r.Options.Scheme)
	if err != nil {
		r.Log.V(2).Info("unable to get gvk for resource, falling back to type")
		return values, nil
	}
	values = append(values,
		"apiVersion", gvk.GroupVersion().String(),
		"kind", gvk.Kind)
	return values, nil
}

func (r *GenericResourceReconciler) resourceLog(desired runtime.Object, details ...interface{}) logr.Logger {
	if len(details) > 0 {
		return r.Log.WithValues(details...)
	}
	return r.Log
}

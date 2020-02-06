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
	"github.com/goph/emperror"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	StateAbsent  StaticDesiredState = "Absent"
	StatePresent StaticDesiredState = "Present"
)

type ResourceBuilders func(object interface{}) []ResourceBuilder
type ResourceBuilder func() (runtime.Object, DesiredState, error)

type DesiredState interface {
	BeforeUpdate(object runtime.Object) error
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

type ReconcilerOpts struct {
	EnableRecreateWorkloadOnImmutableFieldChange     bool
	EnableRecreateWorkloadOnImmutableFieldChangeHelp string
}

// NewReconciler returns GenericResourceReconciler
func NewReconciler(client runtimeClient.Client, log logr.Logger, opts ReconcilerOpts) *GenericResourceReconciler {
	return &GenericResourceReconciler{
		Log:     log,
		Client:  client,
		Options: opts,
	}
}

// CreateResource creates a resource if it doesn't exist
func (r *GenericResourceReconciler) CreateResource(desired runtime.Object) error {
	_, _, err := r.createIfNotExists(desired)
	return err
}

// ReconcileResource reconciles various kubernetes types
func (r *GenericResourceReconciler) ReconcileResource(desired runtime.Object, desiredState DesiredState) (*reconcile.Result, error) {
	key, err := runtimeClient.ObjectKeyFromObject(desired)
	if err != nil {
		return nil, emperror.With(err)
	}
	log := r.Log.WithValues("name", key, "type", reflect.TypeOf(desired))
	debugLog := log.V(1)
	traceLog := log.V(2)
	switch desiredState {
	default:
		created, current, err := r.createIfNotExists(desired)
		if err == nil && created {
			return nil, nil
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create resource %+v", desired)
		}

		// last chance to hook into the desired state armed with the knowledge of the current state
		err = desiredState.BeforeUpdate(current)
		if err != nil {
			return nil, errors.WrapIf(err, "failed to get desired state dynamically")
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
				return nil, errors.Wrap(err, "failed to access resourceVersion from metadata")
			}
			if err := metaAccessor.SetResourceVersion(desired, currentResourceVersion); err != nil {
				return nil, errors.Wrap(err, "failed to set resourceVersion in metadata")
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
							return nil, errors.Wrapf(err, "failed to delete resource %+v", current)
						}
						return &reconcile.Result{
							Requeue:      true,
							RequeueAfter: time.Second * 10,
						}, nil
					} else {
						return nil, errors.New(r.Options.EnableRecreateWorkloadOnImmutableFieldChangeHelp)
					}
				}
				return nil, emperror.WrapWith(err, "updating resource failed")
			}
			debugLog.Info("resource updated")
		}
	case StateAbsent:
		_, err := r.delete(desired)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to delete resource %+v", desired)
		}
	}
	return nil, nil
}

func (r *GenericResourceReconciler) createIfNotExists(desired runtime.Object) (bool, runtime.Object, error) {
	current := reflect.New(reflect.Indirect(reflect.ValueOf(desired)).Type()).Interface().(runtime.Object)
	key, err := runtimeClient.ObjectKeyFromObject(desired)
	if err != nil {
		return false, nil, emperror.With(err)
	}
	log := r.Log.WithValues("name", key, "type", reflect.TypeOf(desired))
	err = r.Client.Get(context.TODO(), key, current)
	if err != nil && !apierrors.IsNotFound(err) {
		return false, nil, emperror.WrapWith(err, "getting resource failed")
	}
	if apierrors.IsNotFound(err) {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(desired); err != nil {
			log.Error(err, "Failed to set last applied annotation", "desired", desired)
		}
		if err := r.Client.Create(context.TODO(), desired); err != nil {
			return false, nil, emperror.WrapWith(err, "creating resource failed")
		}
		switch t := current.DeepCopyObject().(type) {
		case *v1beta1.CustomResourceDefinition:
			err = wait.Poll(time.Second*1, time.Second*10, func() (done bool, err error) {
				err = r.Client.Get(context.TODO(), runtimeClient.ObjectKey{Namespace: t.Namespace, Name: t.Name}, t)
				if err != nil {
					return false, err
				}
				return crdReady(t), nil
			})
			if err != nil {
				return false, nil, errors.WrapIf(err, "failed to wait for the crd to get ready")
			}
		}
		log.Info("resource created")
		return true, current, nil
	}
	log.V(1).Info("resource already exists")
	return false, current, nil
}

func (r *GenericResourceReconciler) delete(desired runtime.Object) (bool, error) {
	key, err := runtimeClient.ObjectKeyFromObject(desired)
	if err != nil {
		return false, emperror.With(err)
	}
	log := r.Log.WithValues("name", key, "type", reflect.TypeOf(desired))
	debugLog := log.V(1)
	current := reflect.New(reflect.Indirect(reflect.ValueOf(desired)).Type()).Interface().(runtime.Object)
	err = r.Client.Get(context.TODO(), key, current)
	if err != nil {
		// If the resource type does not exist we should be ok to move on
		if meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err) {
			return false, nil
		}
		if !apierrors.IsNotFound(err) {
			return false, emperror.WrapWith(err, "getting resource failed")
		} else {
			debugLog.Info("resource not found skipping delete")
			return false, nil
		}
	}
	err = r.Client.Delete(context.TODO(), current)
	if err != nil {
		return false, emperror.With(err)
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
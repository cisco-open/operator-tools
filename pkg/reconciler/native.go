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

	"emperror.dev/errors"
	"github.com/banzaicloud/operator-tools/pkg/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const BanzaiCloudManagedComponent = "banzaicloud.io/managed-component"

type ResourceOwner interface {
	// to be aware of metadata
	metav1.Object
	// to be aware of the owner's type
	runtime.Object
	// control namespace dictates where namespaced objects should belong to
	GetControlNamespace() string
}

type ResourceBuilders func(parent ResourceOwner, object interface{}) []ResourceBuilder
type ResourceBuilder func() (runtime.Object, DesiredState, error)

type NativeReconciledComponent interface {
	ResourceBuilders(parent ResourceOwner, object interface{}) []ResourceBuilder
	RegisterWatches(*builder.Builder)
}

type DefaultReconciledComponent struct {
	builders ResourceBuilders
	watches  func(b *builder.Builder)
}

func NewReconciledComponent(b ResourceBuilders, w func(b *builder.Builder)) NativeReconciledComponent {
	return &DefaultReconciledComponent{
		builders: b,
		watches:  w,
	}
}

func (d *DefaultReconciledComponent) ResourceBuilders(parent ResourceOwner, object interface{}) []ResourceBuilder {
	return d.builders(parent, object)
}

func (d *DefaultReconciledComponent) RegisterWatches(b *builder.Builder) {
	if d.watches != nil {
		d.watches(b)
	}
}

type NativeReconciler struct {
	*GenericResourceReconciler
	client.Client
	reconciledComponent NativeReconciledComponent
	configTranslate     func(runtime.Object) (parent ResourceOwner, config interface{})
	componentName       string
	purgeTypes          []schema.GroupVersionKind
}

func NewNativeReconciler(
	componentName string,
	rec *GenericResourceReconciler,
	client client.Client,
	reconciledComponent NativeReconciledComponent,
	resourceTranslate func(runtime.Object) (parent ResourceOwner, config interface{})) *NativeReconciler {
	return &NativeReconciler{
		GenericResourceReconciler: rec,
		Client:                    client,
		reconciledComponent:       reconciledComponent,
		configTranslate:           resourceTranslate,
		componentName:             componentName,
	}
}

func (rec *NativeReconciler) WithPurgeTypes(purgeTypes []schema.GroupVersionKind) *NativeReconciler {
	rec.purgeTypes = purgeTypes
	return rec
}

func (rec *NativeReconciler) Reconcile(owner runtime.Object) (*reconcile.Result, error) {
	if rec.componentName == "" {
		return nil, errors.New("component name cannot be empty")
	}

	ownerMeta, err := meta.Accessor(owner)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to access owner object meta")
	}

	// generated componentId will be used to purge unwanted objects
	componentId := fmt.Sprintf("%s-%s-%s", ownerMeta.GetName(), ownerMeta.GetUID(), rec.componentName)
	// visited objects wont be purged
	excludeFromPurge := map[string]bool{}

	combinedResult := &CombinedResult{}
	for _, r := range rec.reconciledComponent.ResourceBuilders(rec.configTranslate(owner)) {
		o, state, err := r()
		if err != nil {
			combinedResult.CombineErr(err)
		} else {
			metaObject, err := rec.annotate(o, componentId)
			if err != nil {
				combinedResult.CombineErr(err)
			} else {
				result, err := rec.ReconcileResource(o, state)
				if err == nil {
					excludeFromPurge[utils.ObjectKeyFromObjectMeta(metaObject).String()] = true
				}
				combinedResult.Combine(result, err)
			}
		}
	}
	if combinedResult.Err == nil {
		if err := rec.purge(excludeFromPurge, componentId); err != nil {
			combinedResult.CombineErr(err)
		}
	} else {
		rec.Log.Error(combinedResult.Err, "skip purging results due to previous errors")
	}
	return &combinedResult.Result, combinedResult.Err
}

func (rec *NativeReconciler) purge(excluded map[string]bool, componentId string) error {
	var allErr error
	for _, gvk := range rec.purgeTypes {
		objects := &unstructured.UnstructuredList{}
		objects.SetGroupVersionKind(gvk)
		err := rec.List(context.TODO(), objects)
		if err != nil {
			rec.Log.V(1).Info("retrieving resources to prune type %s: %s not found", gvk.String(), err)
			continue
		}
		for _, o := range objects.Items {
			objectMeta, err := meta.Accessor(&o)
			if err != nil {
				allErr = errors.Combine(allErr, errors.WrapIf(err, "failed to get object metadata"))
				continue
			}
			if excluded[utils.ObjectKeyFromObjectMeta(objectMeta).String()] {
				continue
			}
			if o.GetAnnotations()[BanzaiCloudManagedComponent] == componentId {
				rec.Log.Info("pruning unmmanaged resource",
					"name", objectMeta.GetName(),
					"namespace", objectMeta.GetNamespace(),
					"group", gvk.Group,
					"version", gvk.Version,
					"listKind", gvk.Kind)
				if err := rec.Client.Delete(context.TODO(), &o); err != nil {
					allErr = errors.Combine(allErr, err)
				}
			}
		}
	}
	return allErr
}

func (rec *NativeReconciler) annotate(o runtime.Object, componentId string) (metav1.Object, error) {
	objectMeta, err := meta.Accessor(o)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to access object metadata")
	}
	annotations := objectMeta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if currentComponentId, ok := annotations[BanzaiCloudManagedComponent]; ok {
		if currentComponentId != componentId {
			return nil, errors.Errorf(
				"object actual component id `%s` is different from the one defined by the component `%s`",
				currentComponentId, componentId)
		}
	} else {
		annotations[BanzaiCloudManagedComponent] = componentId
		objectMeta.SetAnnotations(annotations)
	}
	return objectMeta, nil
}

func (rec *NativeReconciler) RegisterWatches(b *builder.Builder) {
	rec.reconciledComponent.RegisterWatches(b)
}

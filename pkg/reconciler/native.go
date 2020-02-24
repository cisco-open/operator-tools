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
	"strings"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	PurgeTypes() []schema.GroupVersionKind
}

type DefaultReconciledComponent struct {
	builders   ResourceBuilders
	watches    func(b *builder.Builder)
	purgeTypes func() []schema.GroupVersionKind
}

func NewReconciledComponent(b ResourceBuilders, w func(b *builder.Builder), p func() []schema.GroupVersionKind) NativeReconciledComponent {
	if p == nil {
		p = func() []schema.GroupVersionKind {
			return nil
		}
	}
	if w == nil {
		w = func(*builder.Builder) {}
	}
	return &DefaultReconciledComponent{
		builders:   b,
		watches:    w,
		purgeTypes: p,
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

func (d *DefaultReconciledComponent) PurgeTypes() []schema.GroupVersionKind {
	return d.purgeTypes()
}

type NativeReconciler struct {
	*GenericResourceReconciler
	client.Client
	scheme              *runtime.Scheme
	reconciledComponent NativeReconciledComponent
	configTranslate     func(runtime.Object) (parent ResourceOwner, config interface{})
	componentName       string
	setControllerRef    bool
}

type NativeReconcilerOpt func(*NativeReconciler)

func NativeReconcilerWithScheme(scheme *runtime.Scheme) NativeReconcilerOpt {
	return func(r *NativeReconciler) {
		r.scheme = scheme
	}
}

func NativeReconcilerSetControllerRef() NativeReconcilerOpt {
	return func(r *NativeReconciler) {
		r.setControllerRef = true
	}
}

func NewNativeReconciler(
	componentName string,
	rec *GenericResourceReconciler,
	client client.Client,
	reconciledComponent NativeReconciledComponent,
	resourceTranslate func(runtime.Object) (parent ResourceOwner, config interface{}),
	opts ...NativeReconcilerOpt) *NativeReconciler {
	reconciler := &NativeReconciler{
		GenericResourceReconciler: rec,
		Client:                    client,
		reconciledComponent:       reconciledComponent,
		configTranslate:           resourceTranslate,
		componentName:             componentName,
	}

	for _, opt := range opts {
		opt(reconciler)
	}

	if reconciler.scheme == nil {
		reconciler.scheme = runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(reconciler.scheme)
	}

	return reconciler
}

func (rec *NativeReconciler) Reconcile(owner runtime.Object) (*reconcile.Result, error) {
	if rec.componentName == "" {
		return nil, errors.New("component name cannot be empty")
	}

	componentID, ownerMeta, err := rec.generateComponentID(owner)
	if err != nil {
		return nil, err
	}
	// visited objects wont be purged
	excludeFromPurge := map[string]bool{}

	combinedResult := &CombinedResult{}
	for _, r := range rec.reconciledComponent.ResourceBuilders(rec.configTranslate(owner)) {
		o, state, err := r()
		if err != nil {
			combinedResult.CombineErr(err)
		} else {
			objectMeta, err := rec.addAnnotation(o, componentID)
			if err != nil {
				combinedResult.CombineErr(err)
				continue
			}
			if rec.setControllerRef {
				if err := controllerutil.SetControllerReference(ownerMeta, objectMeta, rec.scheme); err != nil {
					combinedResult.CombineErr(err)
					continue
				}
			}
			result, err := rec.ReconcileResource(o, state)
			if err == nil {
				resourceID, err := rec.generateResourceID(o)
				if err != nil {
					combinedResult.CombineErr(err)
					continue
				}
				excludeFromPurge[resourceID] = true
			}
			combinedResult.Combine(result, err)
		}
	}
	if combinedResult.Err == nil {
		if err := rec.purge(excludeFromPurge, componentID); err != nil {
			combinedResult.CombineErr(err)
		}
	} else {
		rec.Log.Error(combinedResult.Err, "skip purging results due to previous errors")
	}
	return &combinedResult.Result, combinedResult.Err
}

func (rec *NativeReconciler) generateComponentID(owner runtime.Object) (string, metav1.Object, error) {
	ownerMeta, err := meta.Accessor(owner)
	if err != nil {
		return "", nil, errors.WrapIf(err, "failed to access owner object meta")
	}

	// generated componentId will be used to purge unwanted objects
	identifiers := []string{}
	if ownerMeta.GetName() == "" {
		return "", nil, errors.New("unable to generate component id for resource without a name")
	}
	identifiers = append(identifiers, ownerMeta.GetName())

	if ownerMeta.GetNamespace() != "" {
		identifiers = append(identifiers, ownerMeta.GetNamespace())
	}

	if rec.componentName == "" {
		return "", nil, errors.New("unable to generate component id without a component name")
	}
	identifiers = append(identifiers, rec.componentName)

	gvk, err := apiutil.GVKForObject(owner, rec.scheme)
	if err != nil {
		return "", nil, errors.WrapIf(err, "")
	}
	apiVersion, kind := gvk.ToAPIVersionAndKind()
	identifiers = append(identifiers, apiVersion, strings.ToLower(kind))

	return strings.Join(identifiers, "-"), ownerMeta, nil
}

func (rec *NativeReconciler) generateResourceID(resource runtime.Object) (string, error) {
	resourceMeta, err := meta.Accessor(resource)
	if err != nil {
		return "", errors.WrapIf(err, "failed to access owner object meta")
	}

	// generated componentId will be used to purge unwanted objects
	identifiers := []string{}
	if resourceMeta.GetName() == "" {
		return "", errors.New("unable to generate component id for resource without a name")
	}
	identifiers = append(identifiers, resourceMeta.GetName())

	if resourceMeta.GetNamespace() != "" {
		identifiers = append(identifiers, resourceMeta.GetNamespace())
	}

	gvk, err := apiutil.GVKForObject(resource, rec.scheme)
	if err != nil {
		return "", errors.WrapIf(err, "")
	}
	apiVersion, kind := gvk.ToAPIVersionAndKind()
	identifiers = append(identifiers, apiVersion, strings.ToLower(kind))

	return strings.Join(identifiers, "-"), nil
}

func (rec *NativeReconciler) purge(excluded map[string]bool, componentId string) error {
	var allErr error
	for _, gvk := range rec.reconciledComponent.PurgeTypes() {
		objects := &unstructured.UnstructuredList{}
		objects.SetGroupVersionKind(gvk)
		err := rec.List(context.TODO(), objects)
		if err != nil {
			rec.Log.Error(err, "failed list objects to prune",
				"groupversion", gvk.GroupVersion().String(),
				"kind", gvk.Kind)
			continue
		}
		for _, o := range objects.Items {
			objectMeta, err := meta.Accessor(&o)
			if err != nil {
				allErr = errors.Combine(allErr, errors.WrapIf(err, "failed to get object metadata"))
				continue
			}
			resourceID, err := rec.generateResourceID(&o)
			if err != nil {
				allErr = errors.Combine(allErr, err)
				continue
			}
			if excluded[resourceID] {
				continue
			}
			if objectMeta.GetAnnotations() != nil && objectMeta.GetAnnotations()[BanzaiCloudManagedComponent] == componentId {
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

func (rec *NativeReconciler) addAnnotation(o runtime.Object, componentId string) (metav1.Object, error) {
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

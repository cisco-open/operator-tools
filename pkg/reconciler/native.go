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
	"strings"

	"emperror.dev/errors"
	"github.com/banzaicloud/operator-tools/pkg/resources"
	"github.com/banzaicloud/operator-tools/pkg/wait"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
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

	"github.com/banzaicloud/operator-tools/pkg/types"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

type ResourceOwner interface {
	// to be aware of metadata
	metav1.Object
	// to be aware of the owner's type
	runtime.Object
}

type ResourceOwnerWithControlNamespace interface {
	ResourceOwner
	// control namespace dictates where namespaced objects should belong to
	GetControlNamespace() string
}

type ResourceBuilders func(parent ResourceOwner, object interface{}) []ResourceBuilder
type ResourceBuilder func() (runtime.Object, DesiredState, error)
type ResourceTranslate func(runtime.Object) (parent ResourceOwner, config interface{})
type PurgeTypesFunc func() []schema.GroupVersionKind

func GetResourceBuildersFromObjects(objects []runtime.Object, state DesiredState, modifierFuncs ...resources.ObjectModifierFunc) ([]ResourceBuilder, error) {
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
			ds := DynamicDesiredState{
				DesiredState: state,
			}
			ds.BeforeUpdateFunc = func(current, desired runtime.Object) error {
				for _, f := range []func(current, desired runtime.Object) error{
					ServiceIPModifier,
					KeepLabelsAndAnnotationsModifer,
					KeepServiceAccountTokenReferences,
				} {
					err := f(current, desired)
					if err != nil {
						return err
					}
				}
				return nil
			}

			return o, ds, nil
		})
	}

	return resources, nil
}

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
	scheme                 *runtime.Scheme
	restMapper             meta.RESTMapper
	reconciledComponent    NativeReconciledComponent
	configTranslate        ResourceTranslate
	componentName          string
	setControllerRef       bool
	reconciledObjectStates map[reconciledObjectState][]runtime.Object
	waitBackoff            *wait.Backoff
	objectModifiers        []resources.ObjectModifierWithParentFunc
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

func NativeReconcilerSetRESTMapper(mapper meta.RESTMapper) NativeReconcilerOpt {
	return func(r *NativeReconciler) {
		r.restMapper = mapper
	}
}

func NativeReconcilerWithWait(backoff *wait.Backoff) NativeReconcilerOpt {
	return func(r *NativeReconciler) {
		r.waitBackoff = backoff
	}
}

func NativeReconcilerWithModifier(modifierFunc resources.ObjectModifierWithParentFunc) NativeReconcilerOpt {
	return func(r *NativeReconciler) {
		r.objectModifiers = append(r.objectModifiers, modifierFunc)
	}
}

func NewNativeReconcilerWithDefaults(
	component string,
	client client.Client,
	scheme *runtime.Scheme,
	logger logr.Logger,
	resourceBuilders ResourceBuilders,
	purgeTypes PurgeTypesFunc,
	resourceTranslate ResourceTranslate,
	opts ...NativeReconcilerOpt,
) *NativeReconciler {
	reconcilerOpts := &ReconcilerOpts{
		EnableRecreateWorkloadOnImmutableFieldChange: true,
		Scheme: scheme,
	}

	return NewNativeReconciler(
		component,
		NewGenericReconciler(
			client,
			logger,
			*reconcilerOpts,
		),
		client,
		NewReconciledComponent(
			resourceBuilders,
			nil,
			purgeTypes,
		),
		resourceTranslate,
		opts...,
	)
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

	reconciler.initReconciledObjectStates()

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
LOOP:
	for _, r := range rec.reconciledComponent.ResourceBuilders(rec.configTranslate(owner)) {
		o, state, err := r()
		if err != nil {
			combinedResult.CombineErr(err)
		} else if o == nil || state == nil {
			rec.Log.Info("skipping resource builder reconciliation due to object or desired state was nil")
			continue
		} else {
			var objectMeta metav1.Object
			objectMeta, err = rec.addComponentIDAnnotation(o, componentID)
			if err != nil {
				combinedResult.CombineErr(err)
				continue
			}
			rec.addRelatedToAnnotation(objectMeta, ownerMeta)
			if rec.setControllerRef {
				skipControllerRef := false
				switch o.(type) {
				case *v1beta1.CustomResourceDefinition:
					skipControllerRef = true
				case *v1.CustomResourceDefinition:
					skipControllerRef = true
				case *corev1.Namespace:
					skipControllerRef = true
				}
				if !skipControllerRef {
					// namespaced resource can only own resources in the same namespace
					if ownerMeta.GetNamespace() == "" || ownerMeta.GetNamespace() == objectMeta.GetNamespace() {
						if err := controllerutil.SetControllerReference(ownerMeta, objectMeta, rec.scheme); err != nil {
							combinedResult.CombineErr(err)
							continue
						}
					}
				}
			}

			if len(rec.objectModifiers) > 0 {
				for _, om := range rec.objectModifiers {
					o, err = om(o, owner)
					if err != nil {
						combinedResult.CombineErr(errors.WrapIf(err, "unable to apply object modifier"))
						continue LOOP
					}
				}
			}

			var result *reconcile.Result
			result, err = rec.ReconcileResource(o, state)
			if err == nil {
				resourceID, err := rec.generateResourceID(o)
				if err != nil {
					combinedResult.CombineErr(err)
					continue
				}
				excludeFromPurge[resourceID] = true

				s := ReconciledObjectStatePresent
				if state == StateAbsent {
					s = ReconciledObjectStateAbsent
				}
				rec.addReconciledObjectState(s, o.DeepCopyObject())
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
	if rec.waitBackoff != nil {
		if err := rec.waitForResources(*rec.waitBackoff); err != nil {
			combinedResult.CombineErr(err)
		}
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

func (rec *NativeReconciler) gvkExists(gvk schema.GroupVersionKind) bool {
	if rec.restMapper == nil {
		return true
	}

	mappings, err := rec.restMapper.RESTMappings(gvk.GroupKind(), gvk.Version)
	if apimeta.IsNoMatchError(err) {
		return false
	}
	if err != nil {
		return true
	}

	for _, m := range mappings {
		if gvk == m.GroupVersionKind {
			return true
		}
	}

	return false
}

func (rec *NativeReconciler) purge(excluded map[string]bool, componentId string) error {
	var allErr error
	var purgeObjects []runtime.Object
	for _, gvk := range rec.reconciledComponent.PurgeTypes() {
		rec.Log.V(2).Info("purging GVK", "gvk", gvk)
		if !rec.gvkExists(gvk) {
			continue
		}
		objects := &unstructured.UnstructuredList{}
		objects.SetGroupVersionKind(gvk)
		err := rec.List(context.TODO(), objects)
		if apimeta.IsNoMatchError(err) {
			// skip unknown GVKs
			continue
		}
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
			if objectMeta.GetAnnotations()[types.BanzaiCloudManagedComponent] == componentId {
				rec.Log.Info("will prune unmmanaged resource",
					"name", objectMeta.GetName(),
					"namespace", objectMeta.GetNamespace(),
					"group", gvk.Group,
					"version", gvk.Version,
					"listKind", gvk.Kind)
				purgeObjects = append(purgeObjects, o.DeepCopyObject())
			}
		}
	}

	utils.RuntimeObjects(purgeObjects).Sort(utils.UninstallResourceOrder)
	for _, o := range purgeObjects {
		if err := rec.Client.Delete(context.TODO(), o.(client.Object)); err != nil && !k8serrors.IsNotFound(err) {
			allErr = errors.Combine(allErr, err)
		} else {
			rec.addReconciledObjectState(ReconciledObjectStatePurged, o.DeepCopyObject())
		}
	}
	return allErr
}

type reconciledObjectState string

const (
	ReconciledObjectStateAbsent  reconciledObjectState = "Absent"
	ReconciledObjectStatePresent reconciledObjectState = "Present"
	ReconciledObjectStatePurged  reconciledObjectState = "Purged"
)

func (rec *NativeReconciler) initReconciledObjectStates() {
	rec.reconciledObjectStates = make(map[reconciledObjectState][]runtime.Object)
}

func (rec *NativeReconciler) addReconciledObjectState(state reconciledObjectState, o runtime.Object) {
	rec.reconciledObjectStates[state] = append(rec.reconciledObjectStates[state], o)
}

func (rec *NativeReconciler) GetReconciledObjectWithState(state reconciledObjectState) []runtime.Object {
	return rec.reconciledObjectStates[state]
}

func (rec *NativeReconciler) addRelatedToAnnotation(objectMeta, ownerMeta metav1.Object) {
	annotations := objectMeta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[types.BanzaiCloudRelatedTo] = utils.ObjectKeyFromObjectMeta(ownerMeta).String()
	objectMeta.SetAnnotations(annotations)
}

func (rec *NativeReconciler) addComponentIDAnnotation(o runtime.Object, componentId string) (metav1.Object, error) {
	objectMeta, err := meta.Accessor(o)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to access object metadata")
	}
	annotations := objectMeta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	if currentComponentId, ok := annotations[types.BanzaiCloudManagedComponent]; ok {
		if currentComponentId != componentId {
			return nil, errors.Errorf(
				"object actual component id `%s` is different from the one defined by the component `%s`",
				currentComponentId, componentId)
		}
	} else {
		annotations[types.BanzaiCloudManagedComponent] = componentId
		objectMeta.SetAnnotations(annotations)
	}
	return objectMeta, nil
}

func (rec *NativeReconciler) RegisterWatches(b *builder.Builder) {
	rec.reconciledComponent.RegisterWatches(b)
}

func (rec *NativeReconciler) waitForResources(backoff wait.Backoff) error {
	rcc := wait.NewResourceConditionChecks(rec.Client, backoff, rec.Log, rec.scheme)

	presentObjects := rec.GetReconciledObjectWithState(ReconciledObjectStatePresent)

	err := rcc.WaitForResources("readiness", presentObjects, wait.ExistsConditionCheck, wait.ReadyReplicasConditionCheck)
	if err != nil {
		return err
	}

	absentObjects := append(rec.GetReconciledObjectWithState(ReconciledObjectStateAbsent), rec.GetReconciledObjectWithState(ReconciledObjectStatePurged)...)
	err = rcc.WaitForResources("removal", absentObjects, wait.NonExistsConditionCheck)
	if err != nil {
		return err
	}

	return nil
}

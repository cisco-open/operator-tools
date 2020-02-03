// Copyright © 2020 Banzai Cloud

package reconciler

import (
	"emperror.dev/errors"
	"github.com/banzaicloud/operator-tools/pkg/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NativeReconciledComponent interface {
	ResourceBuilders(object interface{}) []ResourceBuilder
	RegisterWatches(b *builder.Builder)
}

type DefaultReconciledComponent struct {
	Builders ResourceBuilders
	Watches func(b *builder.Builder)
}

func (d *DefaultReconciledComponent) ResourceBuilders(object interface{}) []ResourceBuilder {
	return d.Builders(object)
}

func (d *DefaultReconciledComponent) RegisterWatches(b *builder.Builder) {
	if d.Watches != nil {
		d.Watches(b)
	}
}

type NativeReconciler struct {
	GenericResourceReconciler
	reconciledComponent NativeReconciledComponent
	configTranslate     func(runtime.Object) interface{}
}

func NewNativeReconciler(
	rec GenericResourceReconciler,
	reconciledComponent NativeReconciledComponent,
	resourceTranslate func(runtime.Object) interface{}) *NativeReconciler {
	return &NativeReconciler{
		GenericResourceReconciler: rec,
		reconciledComponent:       reconciledComponent,
		configTranslate:           resourceTranslate,
	}
}

func (rec *NativeReconciler) Reconcile(owner runtime.Object) (*reconcile.Result, error) {
	combinedResult := &CombinedResult{}
	for _, r := range rec.reconciledComponent.ResourceBuilders(rec.configTranslate(owner)) {
		o, state, err := r()
		if err != nil {
			combinedResult.CombineErr(err)
		} else {
			err := rec.metaDecorator(o, owner)
			if err != nil {
				combinedResult.CombineErr(err)
			} else {
				combinedResult.Combine(rec.ReconcileResource(o, state))
			}
		}
	}
	return &combinedResult.Result, combinedResult.Err
}

func (rec *NativeReconciler) RegisterWatches(b *builder.Builder) {
	rec.reconciledComponent.RegisterWatches(b)
}

func (rec *NativeReconciler) metaDecorator(object runtime.Object, owner runtime.Object) error {
	ownerType, err := meta.TypeAccessor(owner)
	if err != nil {
		return errors.Wrapf(err, "failed to access type of owner %+v", owner)
	}

	ownerAccessor, err := meta.Accessor(owner)
	if err != nil {
		return errors.Wrapf(err, "failed to access metadata of owner %+v", owner)
	}

	accessor, err := meta.Accessor(object)
	if err != nil {
		return errors.Wrapf(err, "failed to access meta of object %+v", object)
	}

	ownerRef := metav1.OwnerReference{
		APIVersion: ownerType.GetAPIVersion(),
		Kind:       ownerType.GetKind(),
		Name:       ownerAccessor.GetName(),
		UID:        ownerAccessor.GetUID(),
		Controller: utils.BoolPointer(true),
	}

	refFound := -1
	ownerRefs := accessor.GetOwnerReferences()
	for i, r := range ownerRefs {
		if ownerAccessor.GetUID() == r.UID {
			refFound = i
		}
	}

	if refFound > -1 {
		ownerRefs[refFound] = ownerRef
	} else {
		ownerRefs = append(ownerRefs, ownerRef)
	}
	accessor.SetOwnerReferences(ownerRefs)
	accessor.SetNamespace(ownerAccessor.GetNamespace())

	newLabels := map[string]string{}
	for key, val := range accessor.GetLabels() {
		newLabels[key] = val
	}
	for key, val := range ownerAccessor.GetLabels() {
		newLabels[key] = val
	}
	accessor.SetLabels(newLabels)

	return nil
}

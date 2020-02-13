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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NativeReconciledComponent interface {
	ResourceBuilders(parent metav1.Object, object interface{}) []ResourceBuilder
	RegisterWatches(*builder.Builder)
}

type DefaultReconciledComponent struct {
	Builders ResourceBuilders
	Watches  func(b *builder.Builder)
}

func (d *DefaultReconciledComponent) ResourceBuilders(parent metav1.Object, object interface{}) []ResourceBuilder {
	return d.Builders(parent, object)
}

func (d *DefaultReconciledComponent) RegisterWatches(b *builder.Builder) {
	if d.Watches != nil {
		d.Watches(b)
	}
}

type NativeReconciler struct {
	*GenericResourceReconciler
	reconciledComponent NativeReconciledComponent
	configTranslate     func(runtime.Object) (parent metav1.Object, config interface{})
}

func NewNativeReconciler(
	rec *GenericResourceReconciler,
	reconciledComponent NativeReconciledComponent,
	resourceTranslate func(runtime.Object) (parent metav1.Object, config interface{})) *NativeReconciler {
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
			result, err := rec.ReconcileResource(o, state)
			combinedResult.Combine(result, err)
		}
	}
	return &combinedResult.Result, combinedResult.Err
}

func (rec *NativeReconciler) RegisterWatches(b *builder.Builder) {
	rec.reconciledComponent.RegisterWatches(b)
}
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

package reconciler_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	ottypes "github.com/banzaicloud/operator-tools/pkg/types"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

// FakeResourceOwner object implements the ResourceOwner interface by piggybacking a ConfigMap (oink-oink)
type FakeResourceOwner struct {
	*corev1.ConfigMap
}

func (e *FakeResourceOwner) GetControlNamespace() string {
	return controlNamespace
}

// Assert that a Secret reconciled together with a purged type (ConfigMap) will not get hurt (not even gets dirty)
func TestNativeReconcilerKeepsTheSecret(t *testing.T) {
	nativeReconciler := reconciler.NewNativeReconciler(
		"testcomponent",
		reconciler.NewGenericReconciler(k8sClient, log, reconciler.ReconcilerOpts{}),
		k8sClient,
		reconciler.NewReconciledComponent(
			func(parent reconciler.ResourceOwner, object interface{}) []reconciler.ResourceBuilder {
				parentWithControlNamespace := parent.(reconciler.ResourceOwnerWithControlNamespace)
				rb := []reconciler.ResourceBuilder{}
				// depending on the incoming config we return 0 or more items
				count := cast.ToInt(object)
				for i := 0; i < count; i++ {
					name := fmt.Sprintf("asd-%d", i)
					rb = append(rb, func() (object runtime.Object, state reconciler.DesiredState, e error) {
						return &corev1.ConfigMap{
							ObjectMeta: v1.ObjectMeta{
								Name:      name,
								Namespace: parentWithControlNamespace.GetControlNamespace(),
							},
						}, reconciler.StatePresent, nil
					})
				}
				// this is returned with every call, so it shouldn't change
				rb = append(rb, func() (object runtime.Object, state reconciler.DesiredState, e error) {
					return &corev1.Secret{
						ObjectMeta: v1.ObjectMeta{
							Name:      "keep-the-secret",
							Namespace: parentWithControlNamespace.GetControlNamespace(),
						},
					}, reconciler.StatePresent, nil
				})
				return rb
			},
			func(b *builder.Builder) {},
			func() []schema.GroupVersionKind {
				return []schema.GroupVersionKind{
					{
						Group:   "",
						Version: "v1",
						Kind:    "ConfigMap",
					},
				}
			},
		),
		func(object runtime.Object) (reconciler.ResourceOwner, interface{}) {
			return &FakeResourceOwner{ConfigMap: object.(*corev1.ConfigMap)}, object.(*corev1.ConfigMap).Data["count"]
		},
	)

	fakeOwnerObject := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "example",
			Namespace: controlNamespace,
		},
	}

	setCount := func(c *corev1.ConfigMap, count int) *corev1.ConfigMap {
		if c.Data == nil {
			c.Data = map[string]string{}
		}
		c.Data["count"] = cast.ToString(count)
		return c
	}

	// in the first iteration we create a single configmap and a secret (keep the secret!)

	_, err := nativeReconciler.Reconcile(setCount(fakeOwnerObject, 1))
	if err != nil {
		t.Fatalf("Expected nil, got: %+v", err)
	}

	assertConfigMapList(t, func(l *corev1.ConfigMapList) {
		assert.Len(t, l.Items, 1)
		assert.Equal(t, l.Items[0].Name, "asd-0")
	})
	assertSecretList(t, func(l *corev1.SecretList) {
		assert.Len(t, l.Items, 1)
		assert.Equal(t, l.Items[0].Name, "keep-the-secret")
	})

	assert.Len(t, nativeReconciler.GetReconciledObjectWithState(reconciler.ReconciledObjectStatePurged), 0)

	// next round, the count of configmaps increase to 2, keep the secret!

	_, err = nativeReconciler.Reconcile(setCount(fakeOwnerObject, 2))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	assertConfigMapList(t, func(l *corev1.ConfigMapList) {
		assert.Len(t, l.Items, 2)
		assert.Equal(t, l.Items[0].Name, "asd-0")
		assert.Equal(t, l.Items[1].Name, "asd-1")
	})
	assertSecretList(t, func(l *corev1.SecretList) {
		assert.Len(t, l.Items, 1)
		assert.Equal(t, l.Items[0].Name, "keep-the-secret")
	})

	assert.Len(t, nativeReconciler.GetReconciledObjectWithState(reconciler.ReconciledObjectStatePurged), 0)

	// next round, the count shrinks back to 1, the second configmap should be removed, keep the secret!

	_, err = nativeReconciler.Reconcile(setCount(fakeOwnerObject, 1))
	if err != nil {
		t.Fatalf("Expected nil, got: %+v", err)
	}

	assertConfigMapList(t, func(l *corev1.ConfigMapList) {
		assert.Len(t, l.Items, 1)
		assert.Equal(t, l.Items[0].Name, "asd-0")
	})
	assertSecretList(t, func(l *corev1.SecretList) {
		assert.Len(t, l.Items, 1)
		assert.Equal(t, l.Items[0].Name, "keep-the-secret")
	})

	purged := nativeReconciler.GetReconciledObjectWithState(reconciler.ReconciledObjectStatePurged)
	assert.Len(t, purged, 1)
	assert.Equal(t, purged[0].(*unstructured.Unstructured).GetName(), "asd-1")

	// next round, scale back the configmaps to 0, keep the secret!

	_, err = nativeReconciler.Reconcile(setCount(fakeOwnerObject, 0))
	if err != nil {
		t.Fatalf("Expected nil, got: %+v", err)
	}

	assertConfigMapList(t, func(l *corev1.ConfigMapList) {
		assert.Len(t, l.Items, 0)
	})
	assertSecretList(t, func(l *corev1.SecretList) {
		assert.Len(t, l.Items, 1)
		assert.Equal(t, l.Items[0].Name, "keep-the-secret")
	})

	purged = nativeReconciler.GetReconciledObjectWithState(reconciler.ReconciledObjectStatePurged)
	assert.Len(t, purged, 2)
	assert.Equal(t, purged[0].(*unstructured.Unstructured).GetName(), "asd-1")
	assert.Equal(t, purged[1].(*unstructured.Unstructured).GetName(), "asd-0")
}

func TestNativeReconcilerObjectModifier(t *testing.T) {
	nativeReconciler := createReconcilerForRefTests(
		reconciler.NativeReconcilerWithModifier(func(o, p runtime.Object) (runtime.Object, error) {
			om, _ := meta.Accessor(o)
			pm, _ := meta.Accessor(p)
			om.SetAnnotations(pm.GetAnnotations())
			return o, nil
		}))

	fakeOwnerObject := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "example",
			Namespace: controlNamespace,
			UID:       "something",
			Annotations: map[string]string{
				"parentKey": "parentValue",
			},
		},
	}

	_, err := nativeReconciler.Reconcile(fakeOwnerObject)
	if err != nil {
		t.Fatalf("got error: %s", err.Error())
	}

	assertConfigMapList(t, func(l *corev1.ConfigMapList) {
		assert.Len(t, l.Items, 1)
		assert.Contains(t, l.Items[0].Annotations, "parentKey")
		assert.Equal(t, "parentValue", l.Items[0].Annotations["parentKey"])
	})
}

func TestNativeReconcilerSetNoControllerRefByDefault(t *testing.T) {
	nativeReconciler := createReconcilerForRefTests(
	// without this, controller refs are not going to be applied:
	// reconciler.NativeReconcilerSetControllerRef()
	)

	fakeOwnerObject := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "example",
			Namespace: controlNamespace,
			UID:       "something-fashionable",
		},
	}

	_, err := nativeReconciler.Reconcile(fakeOwnerObject)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assertConfigMapList(t, func(l *corev1.ConfigMapList) {
		assert.Len(t, l.Items, 1)
		assert.Len(t, l.Items[0].OwnerReferences, 0)
	})
}

func TestNativeReconcilerSetControllerRef(t *testing.T) {
	nativeReconciler := createReconcilerForRefTests(
		// without this, controller refs are not going to be applied:
		reconciler.NativeReconcilerSetControllerRef(),
	)

	fakeOwnerObject := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "example",
			Namespace: controlNamespace,
			UID:       "something-fashionable",
		},
	}

	_, err := nativeReconciler.Reconcile(fakeOwnerObject)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	assertConfigMapList(t, func(l *corev1.ConfigMapList) {
		assert.Len(t, l.Items, 1)
		assert.Len(t, l.Items[0].OwnerReferences, 1)
	})
}

func TestNativeReconcilerSetControllerRefMultipleTimes(t *testing.T) {
	nativeReconciler := createReconcilerForRefTests(reconciler.NativeReconcilerSetControllerRef())

	fakeOwnerObject := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "example",
			Namespace: controlNamespace,
			UID:       "something-fashionable",
		},
	}

	for i := 0; i < 2; i++ {
		_, err := nativeReconciler.Reconcile(fakeOwnerObject)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		assertConfigMapList(t, func(l *corev1.ConfigMapList) {
			assert.Len(t, l.Items, 1)
			assert.Len(t, l.Items[0].OwnerReferences, 1)
			assert.Equal(t, fakeOwnerObject.UID, l.Items[0].OwnerReferences[0].UID)
		})
	}
}

func TestNativeReconcilerFailToSetCrossNamespaceControllerRef(t *testing.T) {
	nativeReconciler := createReconcilerForRefTests(reconciler.NativeReconcilerSetControllerRef())

	fakeOwnerObject := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "example",
			Namespace: "another-such-wow-namespace",
			UID:       "something-fashionable",
		},
	}

	_, err := nativeReconciler.Reconcile(fakeOwnerObject)
	if err != nil {
		t.Fatalf("got error: %s", err.Error())
	}
}

func TestCreatedDesiredStateAnnotationWithStaticStatePresent(t *testing.T) {
	desired := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-desired-state-with-static-present",
			Namespace: controlNamespace,
			Annotations: map[string]string{
				ottypes.BanzaiCloudDesiredStateCreated: "true",
			},
		},
		Data: map[string]string{
			"a": "b",
		},
	}

	r := reconciler.NewReconcilerWith(k8sClient)
	result, err := r.ReconcileResource(desired, reconciler.StatePresent)
	if result != nil {
		t.Fatalf("result expected to be nil if everything went smooth")
	}
	if err != nil {
		t.Fatalf("%+v", err)
	}

	desiredMutated := desired.DeepCopy()
	desiredMutated.Data["a"] = "c"

	nr := reconciler.NewNativeReconcilerWithDefaults("test", k8sClient, clientgoscheme.Scheme, log, func(parent reconciler.ResourceOwner, object interface{}) []reconciler.ResourceBuilder {
		return []reconciler.ResourceBuilder{
			func() (runtime.Object, reconciler.DesiredState, error) {
				return desiredMutated, reconciler.StatePresent, nil
			},
		}
	}, func() []schema.GroupVersionKind {
		return nil
	}, func(_ runtime.Object) (reconciler.ResourceOwner, interface{}) {
		return nil, nil
	})

	_, err = nr.Reconcile(desired)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	created := &corev1.ConfigMap{}
	if err := k8sClient.Get(context.TODO(), utils.ObjectKeyFromObjectMeta(desired), created); err != nil {
		t.Fatalf("%+v", err)
	}

	assert.Equal(t, created.Name, desired.Name)
	assert.Equal(t, created.Namespace, desired.Namespace)
	assert.Equal(t, created.Data["a"], desired.Data["a"])
}

func TestCreatedDesiredStateAnnotationWithDynamicStatePresent(t *testing.T) {
	desired := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-desired-state-with-dynamic-present",
			Namespace: controlNamespace,
			Annotations: map[string]string{
				ottypes.BanzaiCloudDesiredStateCreated: "true",
			},
		},
		Data: map[string]string{
			"a": "b",
		},
	}

	r := reconciler.NewReconcilerWith(k8sClient)
	result, err := r.ReconcileResource(desired, reconciler.StatePresent)
	if result != nil {
		t.Fatalf("result expected to be nil if everything went smooth")
	}
	if err != nil {
		t.Fatalf("%+v", err)
	}

	desiredMutated := desired.DeepCopy()
	desiredMutated.Data["a"] = "c"

	nr := reconciler.NewNativeReconcilerWithDefaults("test", k8sClient, clientgoscheme.Scheme, log, func(parent reconciler.ResourceOwner, object interface{}) []reconciler.ResourceBuilder {
		return []reconciler.ResourceBuilder{
			func() (runtime.Object, reconciler.DesiredState, error) {
				return desiredMutated, reconciler.DynamicDesiredState{
					DesiredState: reconciler.StatePresent,
				}, nil
			},
		}
	}, func() []schema.GroupVersionKind {
		return nil
	}, func(_ runtime.Object) (reconciler.ResourceOwner, interface{}) {
		return nil, nil
	})

	_, err = nr.Reconcile(desired)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	created := &corev1.ConfigMap{}
	if err := k8sClient.Get(context.TODO(), utils.ObjectKeyFromObjectMeta(desired), created); err != nil {
		t.Fatalf("%+v", err)
	}

	assert.Equal(t, created.Name, desired.Name)
	assert.Equal(t, created.Namespace, desired.Namespace)
	assert.Equal(t, created.Data["a"], desired.Data["a"])
}

func createReconcilerForRefTests(opts ...reconciler.NativeReconcilerOpt) *reconciler.NativeReconciler {
	return reconciler.NewNativeReconciler(
		"test",
		reconciler.NewGenericReconciler(k8sClient, log, reconciler.ReconcilerOpts{}),
		k8sClient,
		reconciler.NewReconciledComponent(
			func(parent reconciler.ResourceOwner, object interface{}) []reconciler.ResourceBuilder {
				parentWithControlNamespace := parent.(reconciler.ResourceOwnerWithControlNamespace)
				rb := make([]reconciler.ResourceBuilder, 0)
				rb = append(rb, func() (object runtime.Object, state reconciler.DesiredState, e error) {
					return &corev1.ConfigMap{
						ObjectMeta: v1.ObjectMeta{
							Name:      "test-cm",
							Namespace: parentWithControlNamespace.GetControlNamespace(),
						},
					}, reconciler.StatePresent, nil
				})
				return rb
			},
			func(b *builder.Builder) {},
			func() []schema.GroupVersionKind { return []schema.GroupVersionKind{} },
		),
		func(object runtime.Object) (reconciler.ResourceOwner, interface{}) {
			return &FakeResourceOwner{ConfigMap: object.(*corev1.ConfigMap)}, nil
		},
		opts...,
	)
}

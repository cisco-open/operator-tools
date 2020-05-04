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
	"fmt"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/banzaicloud/operator-tools/pkg/reconciler"
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
		reconciler.NewReconciler(k8sClient, log, reconciler.ReconcilerOpts{}),
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

	expectedErrMsg := "cross-namespace owner references are disallowed"

	_, err := nativeReconciler.Reconcile(fakeOwnerObject)
	if err != nil {
		assert.Contains(t, err.Error(), expectedErrMsg)
	} else {
		t.Fatalf("expected: %s", "cross-namespace owner references are disallowed")
	}
}

func createReconcilerForRefTests(opts ...reconciler.NativeReconcilerOpt) *reconciler.NativeReconciler {
	return reconciler.NewNativeReconciler(
		"test",
		reconciler.NewReconciler(k8sClient, log, reconciler.ReconcilerOpts{}),
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

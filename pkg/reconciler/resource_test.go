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
	"testing"

	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestNewReconcilerWith(t *testing.T) {
	desired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: controlNamespace,
		},
		Data: map[string]string{
			"a": "b",
		},
	}
	r := reconciler.NewReconcilerWith(k8sClient, reconciler.WithEnableRecreateWorkload())
	result, err := r.ReconcileResource(desired, reconciler.StatePresent)
	if result != nil {
		t.Fatalf("result expected to be nil if everything went smooth")
	}
	if err != nil {
		t.Fatalf("%+v", err)
	}

	created := &corev1.ConfigMap{}
	if err := k8sClient.Get(context.TODO(), utils.ObjectKeyFromObjectMeta(desired), created); err != nil {
		t.Fatalf("%+v", err)
	}

	assert.Equal(t, created.Name, desired.Name)
	assert.Equal(t, created.Namespace, desired.Namespace)
}

func TestNewReconcilerWithUnstructured(t *testing.T) {
	desired := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": controlNamespace,
			},
			"data": map[string]string{
				"a": "b",
			},
		},
	}
	desired.SetAPIVersion("v1")
	desired.SetKind("ConfigMap")
	r := reconciler.NewReconcilerWith(k8sClient, reconciler.WithEnableRecreateWorkload(), reconciler.WithLog(utils.Log))
	result, err := r.ReconcileResource(desired, reconciler.StatePresent)
	if result != nil {
		t.Fatalf("result expected to be nil if everything went smooth")
	}
	if err != nil {
		t.Fatalf("%+v", err)
	}

	created := &corev1.ConfigMap{}
	if err := k8sClient.Get(context.TODO(), utils.ObjectKeyFromObjectMeta(desired), created); err != nil {
		t.Fatalf("%+v", err)
	}

	assert.Equal(t, created.Name, "test")
	assert.Equal(t, created.Namespace, controlNamespace)
}

func TestRecreateObjectFailIfNotAllowed(t *testing.T) {
	testData := []struct {
		name string
		desired runtime.Object
		reconciler reconciler.ResourceReconciler
		update func(object runtime.Object) runtime.Object
		wantError func(error)
		wantResult func(result *reconcile.Result)
	}{
		{
			name: "fails to recreate service",
			desired: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: testNamespace,
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.100",
					Ports: []corev1.ServicePort{
						{
							Port: 123,
						},
					},
				},
			},
			reconciler: reconciler.NewReconcilerWith(k8sClient,
				reconciler.WithEnableRecreateWorkload(),
				reconciler.WithRecreateEnabledForNothing(),
			),
			update: func(object runtime.Object) runtime.Object {
				object.(*corev1.Service).Spec.ClusterIP = "10.0.0.102"
				return object
			},
			wantError: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "immutable")
			},
		},
		{
			name: "allowed to recreate service by default",
			desired: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: testNamespace,
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.100",
					Ports: []corev1.ServicePort{
						{
							Port: 123,
						},
					},
				},
			},
			reconciler: reconciler.NewReconcilerWith(k8sClient,
				reconciler.WithEnableRecreateWorkload(),
				//reconciler.WithRecreateEnabledForNothing(),
			),
			update: func(object runtime.Object) runtime.Object {
				object.(*corev1.Service).Spec.ClusterIP = "10.0.0.102"
				return object
			},
			wantResult: func(result *reconcile.Result) {
				require.NotNil(t, result)
				require.True(t, result.Requeue)
			},
		},
	}

	for _, tt := range testData {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.reconciler.ReconcileResource(tt.desired, reconciler.StatePresent)
			require.NoError(t, err)

			result, err := tt.reconciler.ReconcileResource(tt.update(tt.desired), reconciler.StatePresent)
			if tt.wantError != nil {
				tt.wantError(err)
			} else {
				require.NoError(t, err)
			}
			if tt.wantResult != nil {
				tt.wantResult(result)
			} else {
				require.Nil(t, result)
			}
		})
	}
}


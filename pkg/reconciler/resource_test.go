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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
		name       string
		desired    runtime.Object
		reconciler reconciler.ResourceReconciler
		update     func(object runtime.Object) runtime.Object
		wantError  func(error)
		wantResult func(result *reconcile.Result)
	}{
		{
			name: "fails to recreate service",
			desired: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test0",
					Namespace: testNamespace,
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.10",
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
				object.(*corev1.Service).Spec.ClusterIP = "10.0.0.11"
				return object
			},
			wantError: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "may not change once set")
			},
		},
		{
			name: "allowed to recreate service by default",
			desired: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: testNamespace,
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.20",
					Ports: []corev1.ServicePort{
						{
							Port: 123,
						},
					},
				},
			},
			reconciler: reconciler.NewReconcilerWith(k8sClient,
				reconciler.WithEnableRecreateWorkload(),
			),
			update: func(object runtime.Object) runtime.Object {
				object.(*corev1.Service).Spec.ClusterIP = "10.0.0.21"
				return object
			},
			wantResult: func(result *reconcile.Result) {
				require.NotNil(t, result)
				require.True(t, result.Requeue)
			},
		},
		{
			name: "recreate service immediately",
			desired: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: testNamespace,
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: "10.0.0.31",
					Ports: []corev1.ServicePort{
						{
							Port: 123,
						},
					},
				},
			},
			reconciler: reconciler.NewReconcilerWith(k8sClient,
				reconciler.WithEnableRecreateWorkload(),
				reconciler.WithRecreateImmediately(),
			),
			update: func(object runtime.Object) runtime.Object {
				object.(*corev1.Service).Spec.ClusterIP = "None"
				object.(*corev1.Service).Spec.ClusterIPs = []string{"None"}
				return object
			},
			wantResult: func(result *reconcile.Result) {
				require.Nil(t, result)
				svc := &corev1.Service{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: testNamespace,
					Name:      "test2",
				}, svc)
				require.NoError(t, err)
				require.Equal(t, svc.Spec.ClusterIP, "None")
			},
		},
		{
			name: "recreate statefulset",
			desired: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: testNamespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "test",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"app": "test",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test",
								},
							},
						},
					},
					ServiceName: "test",
				},
			},
			reconciler: reconciler.NewReconcilerWith(k8sClient,
				reconciler.WithEnableRecreateWorkload(),
				reconciler.WithRecreateErrorMessageCondition(reconciler.MatchImmutableErrorMessages),
				reconciler.WithRecreateImmediately(),
			),
			update: func(object runtime.Object) runtime.Object {
				object.(*appsv1.StatefulSet).Spec.ServiceName = "test2"
				return object
			},
			wantResult: func(result *reconcile.Result) {
				require.Nil(t, result)
				statefulSet := &appsv1.StatefulSet{}
				err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: testNamespace,
					Name:      "test",
				}, statefulSet)
				require.NoError(t, err)
				require.Equal(t, statefulSet.Spec.ServiceName, "test2")
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

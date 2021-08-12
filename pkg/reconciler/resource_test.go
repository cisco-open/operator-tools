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
	"reflect"
	"testing"
	"time"

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

func TestRecreateWorkload(t *testing.T) {
	testData := []struct {
		name string
		desired runtime.Object
		reconciler reconciler.ResourceReconciler
		update func(object runtime.Object) runtime.Object
		wantError func(error)
		wantResult func(result *reconcile.Result)
	}{
		{
			name: "fails to recreate deployment",
			desired: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deploy-0",
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"a": "b",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"a": "b",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									Image: "test",
								},
							},
						},
					},
				},
			},
			reconciler: reconciler.NewReconcilerWith(k8sClient,
				reconciler.WithEnableRecreateWorkload(),
				reconciler.WithRecreateEnabledForNothing(),
			),
			update: func(object runtime.Object) runtime.Object {
				object.(*appsv1.Deployment).Spec.Selector.MatchLabels = map[string]string{
					"c": "d",
				}
				return object
			},
			wantError: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "resource is not allowed to be recreated")
			},
		},
		{
			name: "requeue to recreate deployment",
			desired: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deploy-1",
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"a": "b",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"a": "b",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									Image: "test",
								},
							},
						},
					},
				},
			},
			reconciler: reconciler.NewReconcilerWith(k8sClient,
				reconciler.WithEnableRecreateWorkload(),
			),
			update: func(object runtime.Object) runtime.Object {
				object.(*appsv1.Deployment).Spec.Selector.MatchLabels = map[string]string{
					"c": "d",
				}
				return object
			},
			wantResult: func(result *reconcile.Result) {
				require.Equal(t, &reconcile.Result{
					Requeue:      true,
					RequeueAfter: time.Second * 2,
				}, result)
			},
		},
		{
			name: "delete immediately",
			desired: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deploy-2",
					Namespace: testNamespace,
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"a": "b",
						},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"a": "b",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "test",
									Image: "test",
								},
							},
						},
					},
				},
			},
			reconciler: reconciler.NewReconcilerWith(k8sClient,
				reconciler.WithEnableRecreateWorkload(),
				reconciler.WithRecreateImmediately(),
			),
			update: func(object runtime.Object) runtime.Object {
				newLabels := map[string]string{
					"c": "d",
				}
				object.(*appsv1.Deployment).Spec.Selector.MatchLabels = newLabels
				object.(*appsv1.Deployment).Spec.Template.ObjectMeta.Labels = newLabels
				return object
			},
			wantResult: func(result *reconcile.Result) {
				require.Nil(t, result)
				require.Eventually(t, func() bool {
					deploy := &appsv1.Deployment{}
					err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: testNamespace,
						Name:      "test-deploy-2",
					}, deploy)
					require.NoError(t, err)
					return reflect.DeepEqual(deploy.Spec.Selector.MatchLabels, map[string]string{"c": "d"})
				}, time.Second * 10, time.Second)
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


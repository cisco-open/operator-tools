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

package inventory

import (
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/diff"

	"github.com/cisco-open/operator-tools/pkg/utils"
)

func TestCreateObjectsInventory(t *testing.T) {
	objs := []runtime.Object{
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-ns",
				Name:      "test-svc",
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test-ns",
				Name:      "test-deployment",
			},
			TypeMeta: metav1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "apps/v1",
			},
		},
	}

	cm, _ := CreateObjectsInventory("test-ns", "test-inv", objs)

	expectedConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "test-inv",
		},
		Immutable: utils.BoolPointer(false),
		Data: map[string]string{
			referencesKey: "/v1/Service/test-ns/test-svc,apps/v1/Deployment/test-ns/test-deployment",
		},
	}

	if !reflect.DeepEqual(expectedConfigMap, cm) {
		t.Error(diff.ObjectDiff(expectedConfigMap, cm))
	}
}

func TestGetObjectsFromInventory(t *testing.T) {
	inventory := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-ns",
			Name:      "test-inv",
		},
		Immutable: utils.BoolPointer(false),
		Data: map[string]string{
			referencesKey: "/v1/Service/test-ns/test-svc,apps/v1/Deployment/test-ns/test-deployment",
		},
	}

	objects := GetObjectsFromInventory(inventory)

	expectedSvcObj := &unstructured.Unstructured{}
	expectedSvcObj.SetNamespace("test-ns")
	expectedSvcObj.SetName("test-svc")
	expectedSvcObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	})

	expectedDeplObj := &unstructured.Unstructured{}
	expectedDeplObj.SetNamespace("test-ns")
	expectedDeplObj.SetName("test-deployment")
	expectedDeplObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})

	expectedObjects := []runtime.Object{
		expectedSvcObj,
		expectedDeplObj,
	}

	if !reflect.DeepEqual(expectedObjects, objects) {
		t.Error(diff.ObjectDiff(expectedObjects, objects))
	}
}

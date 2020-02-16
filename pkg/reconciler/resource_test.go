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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

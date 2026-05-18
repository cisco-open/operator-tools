// Copyright © 2026 Banzai Cloud
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
	"encoding/json"
	"testing"

	"github.com/cisco-open/k8s-objectmatcher/patch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/cisco-open/operator-tools/pkg/reconciler"
)

// flattenedMapConfig mimics opentelemetry-operator's AnyConfig: a struct
// whose JSON form is its internal map hoisted to top level. Strategic-merge
// cannot resolve those keys against the Go struct.
type flattenedMapConfig struct {
	Object map[string]any `json:"-"`
}

func (c *flattenedMapConfig) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(c.Object)
}

func (c *flattenedMapConfig) UnmarshalJSON(b []byte) error {
	m := map[string]any{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	c.Object = m
	return nil
}

func (c *flattenedMapConfig) DeepCopyInto(out *flattenedMapConfig) {
	out.Object = make(map[string]any, len(c.Object))
	for k, v := range c.Object {
		out.Object[k] = v
	}
}

type flattenedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              flattenedMapConfig `json:"spec"`
}

func (r *flattenedResource) GetObjectKind() schema.ObjectKind {
	return &r.TypeMeta
}

func (r *flattenedResource) DeepCopyObject() runtime.Object {
	out := &flattenedResource{TypeMeta: r.TypeMeta, ObjectMeta: *r.ObjectMeta.DeepCopy()}
	r.Spec.DeepCopyInto(&out.Spec)
	return out
}

func TestFallbackPatchMaker_FallsBackOnStrategicMergeFailure(t *testing.T) {
	current := &flattenedResource{
		ObjectMeta: metav1.ObjectMeta{Name: "x", Namespace: "default"},
		Spec: flattenedMapConfig{Object: map[string]any{
			"exporters/prometheus": map[string]any{"endpoint": ":9090"},
		}},
	}
	desired := &flattenedResource{
		ObjectMeta: metav1.ObjectMeta{Name: "x", Namespace: "default"},
		Spec: flattenedMapConfig{Object: map[string]any{
			"exporters/prometheus": map[string]any{"endpoint": ":9091"},
		}},
	}

	require.NoError(t, patch.DefaultAnnotator.SetLastAppliedAnnotation(current))

	_, err := patch.DefaultPatchMaker.Calculate(current, desired)
	require.Error(t, err, "expected strategic-merge to fail on flattened-map type")

	maker := reconciler.NewFallbackPatchMaker(patch.DefaultPatchMaker, patch.DefaultAnnotator, &patch.BaseJSONMergePatcher{})
	result, err := maker.Calculate(current, desired)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsEmpty(), "expected non-empty patch for differing specs")
	assert.Contains(t, string(result.Patch), ":9091")
}

func TestFallbackPatchMaker_DelegatesWhenPrimarySucceeds(t *testing.T) {
	current := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "cm", Namespace: "default"},
		Data:       map[string]string{"k": "v1"},
	}
	desired := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "cm", Namespace: "default"},
		Data:       map[string]string{"k": "v2"},
	}
	require.NoError(t, patch.DefaultAnnotator.SetLastAppliedAnnotation(current))

	maker := reconciler.NewFallbackPatchMaker(patch.DefaultPatchMaker, patch.DefaultAnnotator, &patch.BaseJSONMergePatcher{})
	result, err := maker.Calculate(current, desired)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsEmpty())
}

func TestFallbackPatchMaker_EmptyPatchWhenInSync(t *testing.T) {
	obj := &flattenedResource{
		ObjectMeta: metav1.ObjectMeta{Name: "x", Namespace: "default"},
		Spec: flattenedMapConfig{Object: map[string]any{
			"exporters/prometheus": map[string]any{"endpoint": ":9090"},
		}},
	}
	require.NoError(t, patch.DefaultAnnotator.SetLastAppliedAnnotation(obj))

	maker := reconciler.NewFallbackPatchMaker(patch.DefaultPatchMaker, patch.DefaultAnnotator, &patch.BaseJSONMergePatcher{})
	result, err := maker.Calculate(obj, obj)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsEmpty(), "current == desired should produce empty patch")
}

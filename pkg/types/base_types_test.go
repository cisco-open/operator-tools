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

package types_test

import (
	"reflect"
	"testing"

	"github.com/banzaicloud/operator-tools/pkg/types"
	"github.com/banzaicloud/operator-tools/pkg/utils"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMetaBaseEmptyOverrideOnEmptyObject(t *testing.T) {
	original := v1.ObjectMeta{}
	overrides := types.MetaBase{}

	result := overrides.Merge(original)

	if result.Labels != nil {
		t.Error("labels should be nil")
	}

	if result.Annotations != nil {
		t.Error("annotations should be nil")
	}
}

func TestMetaBaseOverrideOnEmptyObject(t *testing.T) {
	original := v1.ObjectMeta{}
	overrides := types.MetaBase{
		Annotations: map[string]string{
			"annotation": "a",
		},
		Labels: map[string]string{
			"label": "l",
		},
	}

	result := overrides.Merge(original)

	if result.Labels["label"] != "l" {
		t.Error("label should be set on empty objectmeta")
	}

	if result.Annotations["annotation"] != "a" {
		t.Error("annotations should be set on empty objectmeta")
	}
}

func TestMetaBaseOverrideOnExistingObject(t *testing.T) {
	original := v1.ObjectMeta{
		Annotations: map[string]string{
			"annotation": "a",
		},
		Labels: map[string]string{
			"label": "l",
		},
	}
	overrides := types.MetaBase{
		Annotations: map[string]string{
			"annotation": "a2",
		},
		Labels: map[string]string{
			"label": "l2",
		},
	}

	result := overrides.Merge(original)

	if result.Labels["label"] != "l2" {
		t.Error("label should be set on empty objectmeta")
	}

	if result.Annotations["annotation"] != "a2" {
		t.Error("annotations should be set on empty objectmeta")
	}
}

func TestDeploymentBaseOverrideOnExistingObject(t *testing.T) {
	original := appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"tik": "tak"},
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "original",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"original", "match", "expression"},
				},
			},
		},
	}
	overrides := types.DeploymentSpecBase{
		Replicas: utils.IntPointer(3),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "another",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"another", "match", "expression"},
				},
			},
		},
		Strategy: &appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		},
	}
	result := overrides.Override(original)

	require.NotNil(t, result)
	require.Equal(t, utils.IntPointer(3), result.Replicas)
	require.Equal(t,
		&metav1.LabelSelector{
			MatchLabels: map[string]string{"tik": "tak", "foo": "bar", "bar": "baz"},
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "original",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"original", "match", "expression"},
				},
				{
					Key:      "another",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"another", "match", "expression"},
				},
			},
		},
		result.Selector)
	require.Equal(t, result.Strategy, appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType})
}

func TestDeploymentBaseOverride(t *testing.T) {
	tests := []struct {
		name string
		base *types.DeploymentSpecBase
		spec appsv1.DeploymentSpec
		want appsv1.DeploymentSpec
	}{
		{
			name: "nil",
			base: nil,
			spec: appsv1.DeploymentSpec{Replicas: utils.IntPointer(77)},
			want: appsv1.DeploymentSpec{Replicas: utils.IntPointer(77)},
		},
		{
			name: "empty",
			base: &types.DeploymentSpecBase{},
			spec: appsv1.DeploymentSpec{},
			want: appsv1.DeploymentSpec{},
		},
		{
			name: "override replicates",
			base: &types.DeploymentSpecBase{Replicas: utils.IntPointer(3)},
			spec: appsv1.DeploymentSpec{},
			want: appsv1.DeploymentSpec{Replicas: utils.IntPointer(3)},
		},
		{
			name: "override deploymentStrategy",
			base: &types.DeploymentSpecBase{
				Strategy: &appsv1.DeploymentStrategy{
					Type: appsv1.RecreateDeploymentStrategyType,
				}},
			spec: appsv1.DeploymentSpec{},
			want: appsv1.DeploymentSpec{
				Strategy: appsv1.DeploymentStrategy{
					Type: appsv1.RecreateDeploymentStrategyType,
				},
			},
		},
		{
			name: "merge matchLabels",
			base: &types.DeploymentSpecBase{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
			},
			spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"tik": "tak"},
				},
			},
			want: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz", "tik": "tak"},
				},
			},
		},
		{
			name: "merge matchExpressions",
			base: &types.DeploymentSpecBase{
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "another",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"another", "match", "expression"},
						},
					},
				},
			},
			spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
				},
			},
			want: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
						{
							Key:      "another",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"another", "match", "expression"},
						},
					},
				},
			},
		},
		{
			name: "merge empty base",
			base: &types.DeploymentSpecBase{},
			spec: appsv1.DeploymentSpec{
				Replicas: utils.IntPointer(77),
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
				Strategy: appsv1.DeploymentStrategy{
					Type: appsv1.RecreateDeploymentStrategyType,
				},
			},
			want: appsv1.DeploymentSpec{
				Replicas: utils.IntPointer(77),
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
				Strategy: appsv1.DeploymentStrategy{
					Type: appsv1.RecreateDeploymentStrategyType,
				},
			},
		},
		{
			name: "merge base to empty spec",
			base: &types.DeploymentSpecBase{
				Replicas: utils.IntPointer(3),
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
				Strategy: &appsv1.DeploymentStrategy{
					Type: appsv1.RecreateDeploymentStrategyType,
				},
			},
			spec: appsv1.DeploymentSpec{},
			want: appsv1.DeploymentSpec{
				Replicas: utils.IntPointer(3),
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
				Strategy: appsv1.DeploymentStrategy{
					Type: appsv1.RecreateDeploymentStrategyType,
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.base.Override(tt.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("base.Override() = \n%#v\nwant\n%#v\n", got, tt.want)
			}
		})
	}

}

func TestStatefulsetBaseOverride(t *testing.T) {
	tests := []struct {
		name string
		base *types.StatefulsetSpecBase
		spec appsv1.StatefulSetSpec
		want appsv1.StatefulSetSpec
	}{
		{
			name: "nil",
			base: nil,
			spec: appsv1.StatefulSetSpec{Replicas: utils.IntPointer(77)},
			want: appsv1.StatefulSetSpec{Replicas: utils.IntPointer(77)},
		},
		{
			name: "empty",
			base: &types.StatefulsetSpecBase{},
			spec: appsv1.StatefulSetSpec{},
			want: appsv1.StatefulSetSpec{},
		},
		{
			name: "override replicates",
			base: &types.StatefulsetSpecBase{Replicas: utils.IntPointer(3)},
			spec: appsv1.StatefulSetSpec{},
			want: appsv1.StatefulSetSpec{Replicas: utils.IntPointer(3)},
		},
		{
			name: "override podManagementPolicy",
			base: &types.StatefulsetSpecBase{PodManagementPolicy: appsv1.ParallelPodManagement},
			spec: appsv1.StatefulSetSpec{PodManagementPolicy: appsv1.OrderedReadyPodManagement},
			want: appsv1.StatefulSetSpec{PodManagementPolicy: appsv1.ParallelPodManagement},
		},
		{
			name: "override updateStrategy",
			base: &types.StatefulsetSpecBase{
				UpdateStrategy: &appsv1.StatefulSetUpdateStrategy{
					Type: appsv1.OnDeleteStatefulSetStrategyType,
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						Partition: utils.IntPointer(33),
					},
				}},
			spec: appsv1.StatefulSetSpec{},
			want: appsv1.StatefulSetSpec{
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
					Type: appsv1.OnDeleteStatefulSetStrategyType,
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						Partition: utils.IntPointer(33),
					},
				},
			},
		},
		{
			name: "merge matchLabels",
			base: &types.StatefulsetSpecBase{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
			},
			spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"tik": "tak"},
				},
			},
			want: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz", "tik": "tak"},
				},
			},
		},
		{
			name: "merge matchExpressions",
			base: &types.StatefulsetSpecBase{
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "another",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"another", "match", "expression"},
						},
					},
				},
			},
			spec: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
				},
			},
			want: appsv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
						{
							Key:      "another",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{"another", "match", "expression"},
						},
					},
				},
			},
		},
		{
			name: "merge empty base",
			base: &types.StatefulsetSpecBase{},
			spec: appsv1.StatefulSetSpec{
				Replicas: utils.IntPointer(77),
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
					Type: appsv1.OnDeleteStatefulSetStrategyType,
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						Partition: utils.IntPointer(33),
					},
				},
			},
			want: appsv1.StatefulSetSpec{
				Replicas: utils.IntPointer(77),
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
					Type: appsv1.OnDeleteStatefulSetStrategyType,
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						Partition: utils.IntPointer(33),
					},
				},
			},
		},
		{
			name: "merge base to empty spec",
			base: &types.StatefulsetSpecBase{
				Replicas: utils.IntPointer(3),
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
				UpdateStrategy: &appsv1.StatefulSetUpdateStrategy{
					Type: appsv1.OnDeleteStatefulSetStrategyType,
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						Partition: utils.IntPointer(33),
					},
				},
			},
			spec: appsv1.StatefulSetSpec{},
			want: appsv1.StatefulSetSpec{
				Replicas: utils.IntPointer(3),
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "original",
							Operator: metav1.LabelSelectorOpNotIn,
							Values:   []string{"original", "match", "expression"},
						},
					},
					MatchLabels: map[string]string{"foo": "bar", "bar": "baz"},
				},
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
					Type: appsv1.OnDeleteStatefulSetStrategyType,
					RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
						Partition: utils.IntPointer(33),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.base.Override(tt.spec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("base.Override() = \n%#v\nwant\n%#v\n", got, tt.want)
			}
		})
	}
}

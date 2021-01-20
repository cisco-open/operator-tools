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

package merge

import (
	"testing"

	"github.com/banzaicloud/operator-tools/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMerge(t *testing.T) {
	base := &v1.DaemonSet{
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-a",
							Image: "image-a",
						},
						{
							Name:  "container-b",
							Image: "image-b",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
						{
							Name:  "container-c",
							Image: "image-c",
						},
					},
				},
			},
		},
	}
	overrides := &v1.DaemonSet{
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-a",
							Image: "image-a-2",
						},
						{
							Name: "container-b",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("123m"),
								},
							},
						},
						{
							Name:  "container-d",
							Image: "image-d",
						},
					},
				},
			},
		},
	}

	result := &v1.DaemonSet{}
	err := Merge(base, overrides, result)
	require.NoError(t, err)

	assert.Len(t, result.Spec.Template.Spec.Containers, 4)

	// container a has a modified image
	assert.Equal(t, "container-a", result.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "image-a-2", result.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, corev1.ResourceRequirements{}, result.Spec.Template.Spec.Containers[0].Resources)

	// container b has the same image but updated resource requirements
	assert.Equal(t, "container-b", result.Spec.Template.Spec.Containers[1].Name)
	assert.Equal(t, "image-b", result.Spec.Template.Spec.Containers[1].Image)
	assert.Equal(t, corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("123m"),
			corev1.ResourceMemory: resource.MustParse("100M"),
		},
	}, result.Spec.Template.Spec.Containers[1].Resources)

	// container d is added as a new item
	assert.Equal(t, base.Spec.Template.Spec.Containers[2], result.Spec.Template.Spec.Containers[3])

	// container c is not modified
	assert.Equal(t, overrides.Spec.Template.Spec.Containers[2], result.Spec.Template.Spec.Containers[2])
}

func TestMergeWithKindredType(t *testing.T) {
	base := &v1.DaemonSet{
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container-a",
							Image: "image-a",
						},
						{
							Name:  "container-b",
							Image: "image-b",
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("100M"),
								},
							},
						},
						{
							Name:  "container-c",
							Image: "image-c",
						},
					},
				},
			},
		},
	}
	overrides := &types.DaemonSetBase{
		Spec: &types.DaemonSetSpecBase{
			Template: &types.PodTemplateBase{
				PodSpec: &types.PodSpecBase{
					Containers: []types.ContainerBase{
						{
							Name:  "container-a",
							Image: "image-a-2",
						},
						{
							Name: "container-b",
							Resources: &corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("123m"),
								},
							},
						},
						{
							Name:  "container-d",
							Image: "image-d",
						},
					},
				},
			},
		},
	}

	result := &v1.DaemonSet{}
	err := Merge(base, overrides, result)
	require.NoError(t, err)

	assert.Len(t, result.Spec.Template.Spec.Containers, 4)

	// container a has a modified image
	assert.Equal(t, "container-a", result.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "image-a-2", result.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, corev1.ResourceRequirements{}, result.Spec.Template.Spec.Containers[0].Resources)

	// container b has the same image but updated resource requirements
	assert.Equal(t, "container-b", result.Spec.Template.Spec.Containers[1].Name)
	assert.Equal(t, "image-b", result.Spec.Template.Spec.Containers[1].Image)
	assert.Equal(t, corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("123m"),
			corev1.ResourceMemory: resource.MustParse("100M"),
		},
	}, result.Spec.Template.Spec.Containers[1].Resources)

	// container d is added as a new item
	assert.Equal(t, base.Spec.Template.Spec.Containers[2], result.Spec.Template.Spec.Containers[3])

	// container c is not modified
	assert.Equal(t, overrides.Spec.Template.PodSpec.Containers[2].Name, result.Spec.Template.Spec.Containers[2].Name)
	assert.Equal(t, overrides.Spec.Template.PodSpec.Containers[2].Image, result.Spec.Template.Spec.Containers[2].Image)
}

func TestMergePrimitiveArrayOverride(t *testing.T) {
	base := &corev1.Service{
		Spec: corev1.ServiceSpec{
			ExternalIPs: []string{
				"a", "b",
			},
		},
	}
	overrides := &types.ServiceBase{
		Spec: &types.ServiceSpecBase{
			ExternalIPs: []string{
				"c", "d",
			},
		},
	}

	result := &corev1.Service{}
	err := Merge(base, overrides, result)
	require.NoError(t, err)

	require.Equal(t, []string{"c", "d"}, result.Spec.ExternalIPs)
}

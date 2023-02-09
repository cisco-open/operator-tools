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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/cisco-open/operator-tools/pkg/typeoverride"
	"github.com/cisco-open/operator-tools/pkg/utils"
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
							// make sure we keep extra fields on the original slice item
							Command: []string{"fake"},
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

	err := Merge(base, overrides)
	require.NoError(t, err)

	assert.Len(t, base.Spec.Template.Spec.Containers, 4)

	// container a has a modified image
	assert.Equal(t, "container-a", base.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "image-a-2", base.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, corev1.ResourceRequirements{}, base.Spec.Template.Spec.Containers[0].Resources)

	// container b has the same image but updated resource requirements
	assert.Equal(t, "container-b", base.Spec.Template.Spec.Containers[1].Name)
	assert.Equal(t, "image-b", base.Spec.Template.Spec.Containers[1].Image)
	assert.Equal(t, corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("123m"),
			corev1.ResourceMemory: resource.MustParse("100M"),
		},
	}, base.Spec.Template.Spec.Containers[1].Resources)

	// container d is added as a new item (note that it will come before container-c)
	assert.Equal(t, corev1.Container{
		Name:  "container-d",
		Image: "image-d",
	}, base.Spec.Template.Spec.Containers[2])

	// container c is not modified (note that it's index will change)
	assert.Equal(t, corev1.Container{
		Name:    "container-c",
		Image:   "image-c",
		Command: []string{"fake"},
	}, base.Spec.Template.Spec.Containers[3])
}

func TestVolume(t *testing.T) {
	base := &v1.DaemonSet{
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container",
							Image: "image",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "asd",
									MountPath: "/asd-path",
									ReadOnly:  true,
								},
							},
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
							Name:  "container",
							Image: "image",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "different",
									MountPath: "/different-path",
								},
							},
						},
					},
				},
			},
		},
	}

	err := Merge(base, overrides)
	require.NoError(t, err)

	require.Equal(t, &v1.DaemonSet{
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container",
							Image: "image",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "different",
									MountPath: "/different-path",
								},
								{
									Name:      "asd",
									MountPath: "/asd-path",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
		},
	}, base)
}

func TestMergeWithEmbeddedType(t *testing.T) {
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
	overrides := &typeoverride.DaemonSet{
		Spec: typeoverride.DaemonSetSpec{
			Template: typeoverride.PodTemplateSpec{
				Spec: typeoverride.PodSpec{
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

	err := Merge(base, overrides)
	require.NoError(t, err)

	assert.Len(t, base.Spec.Template.Spec.Containers, 4)

	// container a has a modified image
	assert.Equal(t, "container-a", base.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "image-a-2", base.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, corev1.ResourceRequirements{}, base.Spec.Template.Spec.Containers[0].Resources)

	// container b has the same image but updated resource requirements
	assert.Equal(t, "container-b", base.Spec.Template.Spec.Containers[1].Name)
	assert.Equal(t, "image-b", base.Spec.Template.Spec.Containers[1].Image)
	assert.Equal(t, corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("123m"),
			corev1.ResourceMemory: resource.MustParse("100M"),
		},
	}, base.Spec.Template.Spec.Containers[1].Resources)

	// container d is added as a new item (note that it will come before container-c)
	assert.Equal(t, corev1.Container{
		Name:  "container-d",
		Image: "image-d",
	}, base.Spec.Template.Spec.Containers[2])

	// container c is not modified (note that it's index will change)
	assert.Equal(t, corev1.Container{
		Name:  "container-c",
		Image: "image-c",
	}, base.Spec.Template.Spec.Containers[3])
}

func TestMergeStatefulSetReplicas(t *testing.T) {
	base := &v1.StatefulSet{
		Spec: v1.StatefulSetSpec{
			Replicas: utils.IntPointer(1),
		},
	}
	overrides := v1.StatefulSet{
		Spec: v1.StatefulSetSpec{
			Replicas: utils.IntPointer(0),
		},
	}
	err := Merge(base, overrides)
	require.NoError(t, err)

	assert.Equal(t, *base.Spec.Replicas, int32(0))
}

func TestMergeArrayOverride(t *testing.T) {
	base := &corev1.Service{
		Spec: corev1.ServiceSpec{
			ExternalIPs: []string{
				"a", "b",
			},
		},
	}
	overrides := &typeoverride.Service{
		Spec: corev1.ServiceSpec{
			ExternalIPs: []string{
				"c", "d",
			},
		},
	}

	err := Merge(base, overrides)
	require.NoError(t, err)

	require.Equal(t, []string{"c", "d"}, base.Spec.ExternalIPs)
}

func TestMergeMap(t *testing.T) {
	base := &corev1.Service{
		ObjectMeta: v12.ObjectMeta{
			Labels: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
	}
	overrides := &corev1.Service{
		ObjectMeta: v12.ObjectMeta{
			Labels: map[string]string{
				"b": "3",
				"c": "4",
			},
		},
	}

	err := Merge(base, overrides)
	require.NoError(t, err)

	require.Equal(t, map[string]string{
		"a": "1",
		"b": "3",
		"c": "4",
	}, base.ObjectMeta.Labels)
}

func TestMergeMapWithEmbeddedType(t *testing.T) {
	base := &corev1.Service{
		ObjectMeta: v12.ObjectMeta{
			Labels: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
	}
	overrides := &typeoverride.Service{
		ObjectMeta: typeoverride.ObjectMeta{
			Labels: map[string]string{
				"b": "3",
				"c": "4",
			},
		},
	}

	err := Merge(base, overrides)
	require.NoError(t, err)

	require.Equal(t, map[string]string{
		"a": "1",
		"b": "3",
		"c": "4",
	}, base.ObjectMeta.Labels)
}

func TestMergeService(t *testing.T) {
	base := &corev1.Service{
		ObjectMeta: v12.ObjectMeta{
			Labels: map[string]string{
				"a": "1",
				"b": "2",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{
				"a": "1",
				"b": "2",
			},
			LoadBalancerIP: "1.2.3.4",
		},
	}
	overrides := &typeoverride.Service{
		ObjectMeta: typeoverride.ObjectMeta{
			Labels: map[string]string{
				"b": "3",
				"c": "4",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http1",
					Port:       80,
					TargetPort: intstr.FromInt(8081),
				},
				{
					Name:       "http2",
					Port:       82,
					TargetPort: intstr.FromInt(8082),
				},
			},
			Selector: map[string]string{
				"b": "3",
				"c": "4",
			},
		},
	}

	err := Merge(base, overrides)
	require.NoError(t, err)

	require.Equal(t, map[string]string{
		"a": "1",
		"b": "3",
		"c": "4",
	}, base.ObjectMeta.Labels)

	require.Equal(t, base.Spec, corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Name:       "http1",
				Port:       80,
				TargetPort: intstr.FromInt(8081),
			},
			{
				Name:       "http2",
				Port:       82,
				TargetPort: intstr.FromInt(8082),
			},
		},
		Selector: map[string]string{
			"a": "1",
			"b": "3",
			"c": "4",
		},
		LoadBalancerIP: "1.2.3.4",
	})
}

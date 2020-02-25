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

package volume

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKubernetesVolume_ApplyVolumeForPodSpec_FailNonExistingContainer(t *testing.T) {
	vol := KubernetesVolume{
		HostPath: &v1.HostPathVolumeSource{
		},
	}

	spec := &v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "test-container",
			},
		},
	}

	err := vol.ApplyVolumeForPodSpec("vol", "none", "/here", spec)
	if err == nil {
		t.Fatalf("should fail with container not found error")
	}

	assert.Contains(t, err.Error(), "failed to find container none")
}

func TestKubernetesVolume_ApplyVolumeForPodSpec(t *testing.T) {
	vol := KubernetesVolume{
		HostPath: &v1.HostPathVolumeSource{
		},
	}

	spec := &v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "test-container",
			},
		},
	}

	if err := vol.ApplyVolumeForPodSpec("vol", "test-container", "/here", spec); err != nil {
		t.Fatalf("+%v", err)
	}

	assert.Equal(t, "vol", spec.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, "/here", spec.Containers[0].VolumeMounts[0].MountPath)
}

func TestKubernetesVolume_ApplyPVCForStatefulSet(t *testing.T) {
	vol := KubernetesVolume{
		PersistentVolumeClaim: &PersistentVolumeClaim{
			PersistentVolumeClaimSpec: v1.PersistentVolumeClaimSpec{
			},
			PersistentVolumeSource: v1.PersistentVolumeClaimVolumeSource{
				ClaimName: "my-claim",
			},
		},
	}

	sts := &appsv1.StatefulSetSpec{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "test-container",
					},
				},
			},
		},
	}

	err := vol.ApplyPVCForStatefulSet("test-container", "/there", sts, func(name string) metav1.ObjectMeta {
		return metav1.ObjectMeta{
			Name: "prefix-" + name,
		}
	})

	assert.Len(t, sts.Template.Spec.Volumes, 0)

	assert.Equal(t, "prefix-my-claim", sts.Template.Spec.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, "/there", sts.Template.Spec.Containers[0].VolumeMounts[0].MountPath)

	assert.Equal(t, "prefix-my-claim", sts.VolumeClaimTemplates[0].Name)

	if err != nil {
		t.Fatalf("+%v", err)
	}
}

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
	corev1 "k8s.io/api/core/v1"
)


//nolint:unused,deadcode
// +docName:"Kubernetes volume abstraction"
// Refers to different types of volumes to be mounted to pods: emptyDir, hostPath, pvc
//
// Leverages core types from kubernetes/api/core/v1
type _docKubernetesVolume interface{}

//nolint:unused,deadcode
// +name:"KubernetesVolume"
// +description:"Kubernetes volume abstraction"
type _metaKubernetesVolume interface{}

// +kubebuilder:object:generate=true

type KubernetesVolume struct {
	// Deprecated, use hostPath
	HostPathLegacy *corev1.HostPathVolumeSource `json:"host_path,omitempty"`
	HostPath       *corev1.HostPathVolumeSource `json:"hostPath,omitempty"`
	EmptyDir       *corev1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// PersistentVolumeClaim defines the Spec and the Source at the same time.
	// The PVC will be created with the configured spec and the name defined in the source.
	PersistentVolumeClaim *PersistentVolumeClaim `json:"pvc,omitempty"`
}

// +kubebuilder:object:generate=true

type PersistentVolumeClaim struct {
	PersistentVolumeClaimSpec corev1.PersistentVolumeClaimSpec         `json:"spec,omitempty"`
	PersistentVolumeSource    corev1.PersistentVolumeClaimVolumeSource `json:"source,omitempty"`
}

// `path` is the path in case the hostPath volume type is used and no path has been defined explicitly
func (v *KubernetesVolume) WithDefaultHostPath(path string) {
	if v.HostPath != nil {
		if v.HostPath.Path == "" {
			v.HostPath.Path = path
		}
	}
}

// GetVolume returns a default emptydir volume if none configured
//
// `name`    will be the name of the volume and the lowest level directory in case a hostPath mount is used
func (v *KubernetesVolume) GetVolume(name string) corev1.Volume {
	volume := corev1.Volume{
		Name: name,
	}
	if v.HostPath != nil {
		volume.VolumeSource = corev1.VolumeSource{
			HostPath: v.HostPath,
		}
		return volume
	} else if v.EmptyDir != nil {
		volume.VolumeSource = corev1.VolumeSource{
			EmptyDir: v.EmptyDir,
		}
		return volume
	} else if v.PersistentVolumeClaim != nil {
		volume.VolumeSource = corev1.VolumeSource{
			PersistentVolumeClaim: &v.PersistentVolumeClaim.PersistentVolumeSource,
		}
		return volume
	}
	// return a default emptydir volume if none configured
	volume.VolumeSource = corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}
	return volume
}

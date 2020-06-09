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

package types

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NameLabel      = "app.kubernetes.io/name"
	InstanceLabel  = "app.kubernetes.io/instance"
	VersionLabel   = "app.kubernetes.io/version"
	ComponentLabel = "app.kubernetes.io/component"
	ManagedByLabel = "app.kubernetes.io/managed-by"
)

// +kubebuilder:object:generate=true

type MetaBase struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// +kubebuilder:object:generate=true

type PodSpecBase struct {
	Tolerations        []corev1.Toleration        `json:"tolerations,omitempty"`
	NodeSelector       map[string]string          `json:"nodeSelector,omitempty"`
	ServiceAccountName string                     `json:"serviceAccountName,omitempty"`
	Affinity           *corev1.Affinity           `json:"affinity,omitempty"`
	SecurityContext    *corev1.PodSecurityContext `json:"securityContext,omitempty"`
	Volumes            []corev1.Volume            `json:"volumes,omitempty"`
	PriorityClassName  string                     `json:"priorityClassName,omitempty"`
}

// +kubebuilder:object:generate=true

type ContainerBase struct {
	Resources       *corev1.ResourceRequirements `json:"resources,omitempty"`
	Image           string                       `json:"image,omitempty"`
	PullPolicy      corev1.PullPolicy            `json:"pullPolicy,omitempty"`
	Command         []string                     `json:"command,omitempty"`
	VolumeMounts    []corev1.VolumeMount         `json:"volumeMounts,omitempty"`
	SecurityContext *corev1.SecurityContext      `json:"securityContext,omitempty"`
}

func (base *ContainerBase) Override(container corev1.Container) corev1.Container {
	if base == nil {
		return container
	}
	if base.Resources != nil {
		container.Resources = *base.Resources
	}
	if base.Image != "" {
		container.Image = base.Image
	}
	if base.PullPolicy != "" {
		container.ImagePullPolicy = base.PullPolicy
	}
	if len(base.Command) > 0 {
		container.Command = base.Command
	}
	if len(base.VolumeMounts) > 0 {
		container.VolumeMounts = base.VolumeMounts
	}
	if base.SecurityContext != nil {
		container.SecurityContext = base.SecurityContext
	}
	return container
}

func (base *MetaBase) Merge(meta v1.ObjectMeta) v1.ObjectMeta {
	if base == nil {
		return meta
	}
	if len(base.Annotations) > 0 {
		if meta.Annotations == nil {
			meta.Annotations = make(map[string]string)
		}
		for key, val := range base.Annotations {
			meta.Annotations[key] = val
		}
	}
	if len(base.Labels) > 0 {
		if meta.Labels == nil {
			meta.Labels = make(map[string]string)
		}
		for key, val := range base.Labels {
			meta.Labels[key] = val
		}
	}
	return meta
}

func (base *PodSpecBase) Override(spec corev1.PodSpec) corev1.PodSpec {
	if base == nil {
		return spec
	}
	if base.SecurityContext != nil {
		spec.SecurityContext = base.SecurityContext
	}
	if base.Tolerations != nil {
		spec.Tolerations = base.Tolerations
	}
	if base.NodeSelector != nil {
		spec.NodeSelector = base.NodeSelector
	}
	if base.ServiceAccountName != "" {
		spec.ServiceAccountName = base.ServiceAccountName
	}
	if base.Affinity != nil {
		spec.Affinity = base.Affinity
	}
	if len(base.Volumes) > 0 {
		spec.Volumes = base.Volumes
	}
	if base.PriorityClassName != "" {
		spec.PriorityClassName = base.PriorityClassName
	}
	return spec
}

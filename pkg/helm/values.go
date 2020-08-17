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

package helm

import (
	corev1 "k8s.io/api/core/v1"
)

// +kubebuilder:object:generate=true
type Image struct {
	Repository string            `json:"repository,omitempty"`
	Tag        string            `json:"tag,omitempty"`
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

type Selectors struct {
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	Affinity     *corev1.Affinity    `json:"affinity,omitempty"`
}

// +kubebuilder:object:generate=true
type EnvironmentVariables struct {
	Env              map[string]string  `json:"env,omitempty"`
	EnvSecrets       []EnvSecret        `json:"envSecrets,omitempty"`
	EnvResourceField []EnvResourceField `json:"envResourceField,omitempty"`
	EnvConfigMap     []EnvConfigMap     `json:"envConfigMaps,omitempty"`
}


// +kubebuilder:object:generate=true
type EnvConfigMap struct {
	Name            string                      `json:"name"`
	ConfigMapKeyRef corev1.ConfigMapKeySelector `json:"configMapKeyRef"`
}

// +kubebuilder:object:generate=true
type EnvResourceField struct {
	Name             string                       `json:"name"`
	ResourceFieldRef corev1.ResourceFieldSelector `json:"resourceFieldRef"`
}

// +kubebuilder:object:generate=true
type EnvSecret struct {
	Name         string                   `json:"name"`
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef"`
}

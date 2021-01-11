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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NameLabel      = "app.kubernetes.io/name"
	InstanceLabel  = "app.kubernetes.io/instance"
	VersionLabel   = "app.kubernetes.io/version"
	ComponentLabel = "app.kubernetes.io/component"
	ManagedByLabel = "app.kubernetes.io/managed-by"

	BanzaiCloudManagedComponent = "banzaicloud.io/managed-component"
	BanzaiCloudOwnedBy          = "banzaicloud.io/owned-by"
	BanzaiCloudRelatedTo        = "banzaicloud.io/related-to"
)

type ObjectKey struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type ReconcileStatus string

const (
	// Used for components and for aggregated status
	ReconcileStatusFailed ReconcileStatus = "Failed"

	// Used for components and for aggregated status
	ReconcileStatusReconciling ReconcileStatus = "Reconciling"

	// Used for components
	ReconcileStatusAvailable ReconcileStatus = "Available"
	ReconcileStatusUnmanaged ReconcileStatus = "Unmanaged"
	ReconcileStatusRemoved   ReconcileStatus = "Removed"

	// Used for aggregated status if all the components are stableized (Available, Unmanaged or Removed)
	ReconcileStatusSucceeded ReconcileStatus = "Succeeded"

	// Used to trigger reconciliation for a resource that otherwise ignores status changes, but listens to the Pending state
	// See PendingStatusPredicate in pkg/reconciler
	ReconcileStatusPending ReconcileStatus = "Pending"
)

func (s ReconcileStatus) Stable() bool {
	return s == ReconcileStatusUnmanaged || s == ReconcileStatusRemoved || s == ReconcileStatusAvailable
}

func (s ReconcileStatus) Available() bool {
	return s == ReconcileStatusAvailable || s == ReconcileStatusSucceeded
}

func (s ReconcileStatus) Failed() bool {
	return s == ReconcileStatusFailed
}

func (s ReconcileStatus) Pending() bool {
	return s == ReconcileStatusReconciling || s == ReconcileStatusPending
}

// Computes an aggregated state based on component statuses
func AggregatedState(componentStatuses []ReconcileStatus) ReconcileStatus {
	overallStatus := ReconcileStatusReconciling
	statusMap := make(map[ReconcileStatus]bool)
	hasUnstable := false
	for _, cs := range componentStatuses {
		if cs != "" {
			statusMap[cs] = true
		}
		if !(cs == "" || cs.Stable()) {
			hasUnstable = true
		}
	}

	if statusMap[ReconcileStatusFailed] {
		overallStatus = ReconcileStatusFailed
	} else if statusMap[ReconcileStatusReconciling] {
		overallStatus = ReconcileStatusReconciling
	}

	if !hasUnstable {
		overallStatus = ReconcileStatusSucceeded
	}
	return overallStatus
}

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

func (base *MetaBase) Merge(meta metav1.ObjectMeta) metav1.ObjectMeta {
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

// +kubebuilder:object:generate=true

type DeploymentSpecBase struct {
	Replicas *int32                     `json:"replicas,omitempty"`
	Selector *metav1.LabelSelector      `json:"selector,omitempty"`
	Strategy *appsv1.DeploymentStrategy `json:"strategy,omitempty"`
}

func (base *DeploymentSpecBase) Override(spec appsv1.DeploymentSpec) appsv1.DeploymentSpec {
	if base == nil {
		return spec
	}
	if base.Replicas != nil {
		spec.Replicas = base.Replicas
	}
	spec.Selector = mergeSelectors(base.Selector, spec.Selector)
	if base.Strategy != nil {
		spec.Strategy = *base.Strategy
	}
	return spec
}

// +kubebuilder:object:generate=true

type StatefulsetSpecBase struct {
	Replicas            *int32                            `json:"replicas,omitempty"`
	Selector            *metav1.LabelSelector             `json:"selector,omitempty"`
	PodManagementPolicy appsv1.PodManagementPolicyType    `json:"podManagementPolicy,omitempty"`
	UpdateStrategy      *appsv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty"`
}

func (base *StatefulsetSpecBase) Override(spec appsv1.StatefulSetSpec) appsv1.StatefulSetSpec {
	if base == nil {
		return spec
	}
	if base.Replicas != nil {
		spec.Replicas = base.Replicas
	}
	spec.Selector = mergeSelectors(base.Selector, spec.Selector)
	if base.PodManagementPolicy != "" {
		spec.PodManagementPolicy = base.PodManagementPolicy
	}
	if base.UpdateStrategy != nil {
		spec.UpdateStrategy = *base.UpdateStrategy

	}

	return spec
}

// +kubebuilder:object:generate=true

type DaemonSetSpecBase struct {
	Selector             *metav1.LabelSelector           `json:"selector,omitempty"`
	UpdateStrategy       *appsv1.DaemonSetUpdateStrategy `json:"updateStrategy,omitempty"`
	MinReadySeconds      int32                           `json:"minReadySeconds,omitempty"`
	RevisionHistoryLimit *int32                          `json:"revisionHistoryLimit,omitempty"`
}

func (base *DaemonSetSpecBase) Override(spec appsv1.DaemonSetSpec) appsv1.DaemonSetSpec {
	if base == nil {
		return spec
	}

	spec.Selector = mergeSelectors(base.Selector, spec.Selector)
	if base.UpdateStrategy != nil {
		spec.UpdateStrategy = *base.UpdateStrategy

	}
	if base.MinReadySeconds != 0 {
		spec.MinReadySeconds = base.MinReadySeconds

	}
	if base.RevisionHistoryLimit != nil {
		spec.RevisionHistoryLimit = base.RevisionHistoryLimit

	}
	return spec
}

func mergeSelectors(base, spec *metav1.LabelSelector) *metav1.LabelSelector {
	if base == nil {
		return spec
	}

	if base.MatchLabels != nil {
		if spec == nil {
			spec = &metav1.LabelSelector{}
		}
		if spec.MatchLabels == nil {
			spec.MatchLabels = make(map[string]string)
		}
		for k, v := range base.MatchLabels {
			spec.MatchLabels[k] = v
		}
	}
	if base.MatchExpressions != nil {
		spec.MatchExpressions = append(spec.MatchExpressions, base.MatchExpressions...)
	}

	return spec
}

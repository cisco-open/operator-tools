// Copyright © 2021 Banzai Cloud
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

package reconciler

import (
	"regexp"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RecreateConfig struct {
	Delete              bool
	RecreateImmediately bool
	DeletePropagation   metav1.DeletionPropagation
	Delay               int32
}

type RecreateResourceCondition func(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error)

type FirstMatchingResourceCondition struct {
	Conditions []RecreateResourceCondition
}

func (cond FirstMatchingResourceCondition) Condition(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error) {
	for _, c := range cond.Conditions {
		config, err := c(kind, status)
		if err != nil {
			return config, err
		}
		// skip to the next condition
		if config.Delete == false {
			continue
		}
		return config, nil
	}
	return RecreateConfig{}, nil
}

// Change in the workload (deployment, statefulset, daemonset) selector results in the following error:
// "spec.selector [...] field is immutable"
func ImmutableFieldChangeCondition(config RecreateConfig) func(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error) {
	return func(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error) {
		found := false
		for _, gk := range DefaultRecreateEnabledGroupKinds {
			if gk == kind.GroupKind() {
				found = true
			}
		}
		if !found {
			return RecreateConfig{}, nil
		}
		match, err := regexp.Match(`field is immutable`, []byte(status.Message))
		if err != nil {
			return RecreateConfig{}, err
		}
		if match {
			return config, nil
		}
		return RecreateConfig{}, nil
	}
}

func StatefulSetFieldChangeCondition(config RecreateConfig) func(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error) {
	return func(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error) {
		if strings.Contains(status.Message, "Forbidden: updates to statefulset spec for fields other than") {
			return config, nil
		}
		return RecreateConfig{}, nil
	}
}

// Use this option for the legacy behaviour
func WithRecreateEnabledFor(condition RecreateResourceCondition) ResourceReconcilerOption {
	return func(o *ReconcilerOpts) {
		o.RecreateEnabledResourceCondition = condition
	}
}

func WithEnableRecreateWorkload() ResourceReconcilerOption {
	return func(o *ReconcilerOpts) {
		o.EnableRecreateWorkloadOnImmutableFieldChange = true
	}
}

// Use this option for the legacy behaviour
func WithRecreateEnabledForAll() ResourceReconcilerOption {
	return func(o *ReconcilerOpts) {
		o.RecreateEnabledResourceCondition = func(_ schema.GroupVersionKind, _ metav1.Status) (RecreateConfig, error) {
			return RecreateConfig{
				Delete:              true,
				RecreateImmediately: false,
				DeletePropagation:   metav1.DeletePropagationForeground,
				Delay:               DefaultRecreateRequeueDelay,
			}, nil
		}
	}
}

// Matches no GVK
func WithRecreateEnabledForNothing() ResourceReconcilerOption {
	return func(o *ReconcilerOpts) {
		o.RecreateEnabledResourceCondition = func(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error) {
			return RecreateConfig{}, nil
		}
	}
}

// Delete workloads immediately without waiting for dependents to get GCd
func WithRecreateImmediately() ResourceReconcilerOption {
	return func(o *ReconcilerOpts) {
		o.RecreateEnabledResourceCondition = func(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error) {
			return RecreateConfig{
				RecreateImmediately: true,
				Delete:              true,
				DeletePropagation:   metav1.DeletePropagationBackground,
				Delay:               DefaultRecreateRequeueDelay,
			}, nil
		}
	}
}

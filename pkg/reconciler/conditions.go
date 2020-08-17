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

package reconciler

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/operator-tools/pkg/wait"
)

var DefaultBackoff = wait.Backoff{
	Duration: time.Second * 5,
	Factor:   1,
	Jitter:   0,
	Steps:    12,
}

type ConditionType string

type ObjectKeyWithGVK struct {
	ObjectKey client.ObjectKey        `json:"objectKey,omitempty"`
	GVK       schema.GroupVersionKind `json:"gvk,omitempty"`
}

type ResourceCondition struct {
	ID           string                              `json:"id,omitempty"`
	Description  string                              `json:"shortDescription,omitempty"`
	Checks       []wait.ResourceConditionCheck       `json:"checks,omitempty"`
	CustomChecks []wait.CustomResourceConditionCheck `json:"customChecks,omitempty"`
	Object       *ObjectKeyWithGVK                   `json:"object,omitempty"`
}

func (rec *Reconciler) CheckResourceConditions(conditions []ResourceCondition, backoff *wait.Backoff) error {
	if backoff == nil {
		backoff = &DefaultBackoff
	}

	log := rec.log.WithName("pre-checks")
	c := wait.NewResourceConditionChecks(rec.mgr.GetClient(), *backoff, log, rec.mgr.GetScheme())

	log.Info("checking resource pre-conditions")

	for _, condition := range conditions {
		if condition.Object != nil && len(condition.Checks) > 0 {
			o, err := rec.mgr.GetScheme().New(condition.Object.GVK)
			if err != nil {
				return err
			}
			if mo, ok := o.(metav1.Object); ok {
				mo.SetName(condition.Object.ObjectKey.Name)
				mo.SetNamespace(condition.Object.ObjectKey.Namespace)
			}
			err = c.WaitForResources(condition.ID, []runtime.Object{o}, condition.Checks...)
			if err != nil {
				return err
			}
		}

		if len(condition.CustomChecks) > 0 {
			err := c.WaitForCustomConditionChecks(condition.ID, condition.CustomChecks...)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

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

package wait

import (
	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceConditionCheck func(runtime.Object, error) bool
type CustomResourceConditionCheck func() (bool, error)

func ExistsConditionCheck(obj runtime.Object, k8serror error) bool {
	return k8serror == nil
}

func NonExistsConditionCheck(obj runtime.Object, k8serror error) bool {
	return k8serrors.IsNotFound(k8serror) || meta.IsNoMatchError(k8serror)
}

func CRDEstablishedConditionCheck(obj runtime.Object, k8serror error) bool {
	var resource *apiextensionsv1beta1.CustomResourceDefinition
	var ok bool
	if resource, ok = obj.(*apiextensionsv1beta1.CustomResourceDefinition); ok {
		for _, condition := range resource.Status.Conditions {
			if condition.Type == apiextensionsv1beta1.Established {
				if condition.Status == apiextensionsv1beta1.ConditionTrue {
					return true
				}
			}
		}
		return false
	}

	var resourcev1 *apiextensionsv1.CustomResourceDefinition
	if resourcev1, ok = obj.(*apiextensionsv1.CustomResourceDefinition); ok {
		for _, condition := range resourcev1.Status.Conditions {
			if condition.Type == apiextensionsv1.Established {
				if condition.Status == apiextensionsv1.ConditionTrue {
					return true
				}
			}
		}
		return false
	}

	return true
}

func ReadyReplicasConditionCheck(obj runtime.Object, k8serror error) bool {
	switch o := obj.(type) {
	case *appsv1.Deployment:
		return o.Status.ReadyReplicas == o.Status.Replicas
	case *appsv1.StatefulSet:
		return o.Status.ReadyReplicas == o.Status.Replicas
	case *appsv1.DaemonSet:
		return o.Status.DesiredNumberScheduled == o.Status.NumberReady
	default:
		// return true for unconvertable objects
		return true
	}
}

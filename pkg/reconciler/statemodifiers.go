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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/banzaicloud/operator-tools/pkg/types"
)

func ServiceIPModifier(current, desired runtime.Object) error {
	if co, ok := current.(*corev1.Service); ok {
		do := desired.(*corev1.Service)
		do.Spec.ClusterIP = co.Spec.ClusterIP
	}

	return nil
}

func KeepLabelsAndAnnotationsModifer(current, desired runtime.Object) error {
	if desiredMetaObject, ok := desired.(metav1.Object); ok {
		base := types.MetaBase{
			Annotations: desiredMetaObject.GetAnnotations(),
			Labels:      desiredMetaObject.GetLabels(),
		}
		if metaObject, ok := current.DeepCopyObject().(metav1.Object); ok {
			merged := base.Merge(metav1.ObjectMeta{
				Labels:      metaObject.GetLabels(),
				Annotations: metaObject.GetAnnotations(),
			})
			desiredMetaObject.SetAnnotations(merged.Annotations)
			desiredMetaObject.SetLabels(merged.Labels)
		}
	}

	return nil
}

func KeepServiceAccountTokenReferences(current, desired runtime.Object) error {
	if co, ok := current.(*corev1.ServiceAccount); ok {
		do := desired.(*corev1.ServiceAccount)
		do.Secrets = co.Secrets
	}

	return nil
}

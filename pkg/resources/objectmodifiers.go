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

package resources

import (
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ObjectModifierFunc func(o runtime.Object) (runtime.Object, error)

var DefaultModifiers = []ObjectModifierFunc{
	ClearCRDStatusModifier,
	ClusterScopeNamespaceFixModifier,
	MutatingWebhookConfigurationModifier,
	ValidatingWebhookConfigurationModifier,
}

func ClearCRDStatusModifier(o runtime.Object) (runtime.Object, error) {
	if crd, ok := o.(*apiextensionsv1beta1.CustomResourceDefinition); ok {
		crd.Status = apiextensionsv1beta1.CustomResourceDefinitionStatus{}
	}

	return o, nil
}

func ClusterScopeNamespaceFixModifier(o runtime.Object) (runtime.Object, error) {
	if obj, ok := o.(*policyv1beta1.PodSecurityPolicy); ok {
		obj.Namespace = ""
	}

	return o, nil
}

func MutatingWebhookConfigurationModifier(o runtime.Object) (runtime.Object, error) {
	if obj, ok := o.(*admissionregistrationv1beta1.MutatingWebhookConfiguration); ok {
		allScope := admissionregistrationv1beta1.AllScopes
		for i, wh := range obj.Webhooks {
			for l, r := range wh.Rules {
				if r.Scope == nil {
					r.Scope = &allScope
					wh.Rules[l] = r
				}
			}
			obj.Webhooks[i] = wh
		}
	}

	return o, nil
}

func ValidatingWebhookConfigurationModifier(o runtime.Object) (runtime.Object, error) {
	if obj, ok := o.(*admissionregistrationv1beta1.ValidatingWebhookConfiguration); ok {
		allScope := admissionregistrationv1beta1.AllScopes
		for i, wh := range obj.Webhooks {
			for l, r := range wh.Rules {
				if r.Scope == nil {
					r.Scope = &allScope
					wh.Rules[l] = r
				}
			}
			obj.Webhooks[i] = wh
		}
	}

	return o, nil
}

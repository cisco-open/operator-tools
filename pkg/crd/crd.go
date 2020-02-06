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

package crd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CRD struct {
	sri ServerResourcesInterface
}

func NewCRD(sri ServerResourcesInterface) *CRD {
	return &CRD{
		sri: sri,
	}
}

// ListAPIResources returns all API resources registered for a given API group and version.
func (c *CRD) ListAPIResources(groupVersion metav1.GroupVersion) ([]metav1.APIResource, error) {
	resp, err := c.sri.ServerResourcesForGroupVersion(groupVersion.String())
	if err != nil {
		return nil, err
	}

	return resp.APIResources, nil
}

// HasAPIResource checks whether a given resource is registered under a given API group and resource or not.
func (c *CRD) HasAPIResource(groupVersion metav1.GroupVersion, name string) (bool, error) {
	resources, err := c.ListAPIResources(groupVersion)
	if err != nil {
		return false, err
	}

	for _, res := range resources {
		if res.Name == name {
			return true, nil
		}
	}

	return false, nil
}

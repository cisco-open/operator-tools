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

package inventory

import (
	"sync"

	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
)

var staticResourceScope map[string]bool
var dynamicResourceScope map[string]bool

var mutex = sync.RWMutex{}

func AddStaticResourceScope(gk schema.GroupKind, namespaced bool) {
	mutex.Lock()
	defer mutex.Unlock()

	if staticResourceScope == nil {
		staticResourceScope = make(map[string]bool)
	}
	staticResourceScope[gk.String()] = namespaced
}

func getStaticResourceScope(gk schema.GroupKind) (bool, bool) {
	mutex.RLock()
	defer mutex.RUnlock()

	namespaced, ok := staticResourceScope[gk.String()]
	return namespaced, ok
}

// initializeAPIResources discovers api resources and returns true if initialization actually happened
func initializeAPIResources(discoveryClient discovery.DiscoveryInterface) (bool, error) {
	if len(dynamicResourceScope) == 0 {
		return true, discoverAPIResources(discoveryClient)
	}
	return false, nil
}

func discoverAPIResources(discoveryClient discovery.DiscoveryInterface) error {
	mutex.Lock()
	defer mutex.Unlock()

	_, apiResourcesList, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return errors.WrapIf(err, "couldn't retrieve the list of resources supported by API server")
	}
	for _, apiResources := range apiResourcesList {
		if apiResources != nil {
			gv, err := schema.ParseGroupVersion(apiResources.GroupVersion)
			if err != nil {
				return errors.WrapIff(err, "unable to parse groupversion %s", apiResources.GroupVersion)
			}
			for _, apiResource := range apiResources.APIResources {
				gk := schema.GroupKind{Group: gv.Group, Kind: apiResource.Kind}
				if dynamicResourceScope == nil {
					dynamicResourceScope = make(map[string]bool)
				}
				dynamicResourceScope[gk.String()] = apiResource.Namespaced
			}
		}
	}
	return nil
}

// getDynamicResourceScope returns
func getDynamicResourceScope(gk schema.GroupKind) (bool, bool) {
	mutex.RLock()
	defer mutex.RUnlock()

	namespaced, ok := dynamicResourceScope[gk.String()]
	return namespaced, ok
}
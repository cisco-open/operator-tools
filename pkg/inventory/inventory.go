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
	"context"
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

const (
	CustomResourceDefinition = "CustomResourceDefinition"
	Namespace                = "Namespace"
)

// State holder between reconcile phases
type InventoryData struct {
	ObjectsToDelete []runtime.Object
	CurrentObjects  []runtime.Object
	DesiredObjects  []runtime.Object
}

// A generalized structure to enable us to attach additional inventory management
// around the native reconcile loop and adds the capability to store state between
// reconcile phases (see operator-tool's NativeReconciler)
type Inventory struct {
	genericClient client.Client
	log           logr.Logger
	inventoryData InventoryData

	// map of GVK of cluster scoped API resources
	clusterScopedAPIResources map[string]struct{}

	// Discovery client to look up API resources
	discoveryClient discovery.DiscoveryInterface
}

func NewInventory(client client.Client, log logr.Logger, clusterResources map[string]struct{}) (*Inventory, error) {
	if clusterResources == nil {
		return nil, errors.New("list of cluster scoped resources is required")
	}
	return &Inventory{
		genericClient:             client,
		log:                       log,
		clusterScopedAPIResources: clusterResources,
	}, nil
}

func NewDiscoveryInventory(client client.Client, log logr.Logger, discovery discovery.DiscoveryInterface) *Inventory {
	return &Inventory{
		genericClient:   client,
		log:             log,
		discoveryClient: discovery,
	}
}

func CreateObjectsInventory(namespace, name string, objects []runtime.Object) (*core.ConfigMap, error) {
	resourceURLs := make([]string, len(objects))
	for i := range objects {
		objMeta, err := meta.Accessor(objects[i])
		if err != nil {
			return nil, err
		}
		objGVK := objects[i].GetObjectKind().GroupVersionKind()
		resourceURLs[i] = fmt.Sprintf("%s/%s/%s/%s/%s",
			objGVK.Group,
			objGVK.Version,
			objGVK.Kind,
			objMeta.GetNamespace(),
			objMeta.GetName())
	}
	cm := core.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Immutable: utils.BoolPointer(false),
		Data: map[string]string{
			"refs": strings.Join(resourceURLs, ","),
		},
	}

	return &cm, nil
}

func GetObjectsFromInventory(inventory core.ConfigMap) (objects []runtime.Object) {
	resourceURLs := strings.Split(inventory.Data["refs"], ",")

	for i := range resourceURLs {
		if resourceURLs[i] == "" {
			continue
		}
		parts := strings.Split(resourceURLs[i], "/")

		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   parts[0],
			Version: parts[1],
			Kind:    parts[2],
		})
		u.SetNamespace(parts[3])
		u.SetName(parts[4])

		objects = append(objects, u)
	}

	return
}

// Hand over a GVK list to the native reconcile loop's purge method
func (c *Inventory) TypesToPurge() []schema.GroupVersionKind {
	currentObjects := c.inventoryData.ObjectsToDelete
	groupVersionKindDict := make(map[schema.GroupVersionKind]struct{})

	for _, currentObject := range currentObjects {
		gvk := currentObject.GetObjectKind().GroupVersionKind()

		if gvk.Kind == CustomResourceDefinition || gvk.Kind == Namespace {
			continue
		}

		if objMeta, err := meta.Accessor(currentObject); err == nil {
			c.log.V(1).Info("mark object for deletion", "gvk", gvk.String(), "namespace", objMeta.GetNamespace(), "name", objMeta.GetName())
			groupVersionKindDict[gvk] = struct{}{}
		}
	}

	groupVersionKindList := make([]schema.GroupVersionKind, 0, len(groupVersionKindDict))
	for k := range groupVersionKindDict {
		groupVersionKindList = append(groupVersionKindList, k)
	}

	return groupVersionKindList
}

// Fetch list of resources made by the previous reconcile loop and store into an attached context
// Return a new list of resources which will be reconciled among with the other resources we listed here
func (c *Inventory) PrepareDesiredObjects(ns, componentName string, parent reconciler.ResourceOwner, resourceBuilders []reconciler.ResourceBuilder) (*core.ConfigMap, error) {
	var err error
	var desiredObjects []runtime.Object
	var objectsInventory core.ConfigMap
	objectsInventoryName := fmt.Sprintf("%s-%s-%s-object-inventory", parent.GetName(), ns, componentName)

	// collect
	err = c.genericClient.Get(context.TODO(), types.NamespacedName{Namespace: ns, Name: objectsInventoryName}, &objectsInventory)
	if err != nil && !meta.IsNoMatchError(err) && !apierrors.IsNotFound(err) {
		return nil, errors.WrapIfWithDetails(err,
			"during object inventory fetch...",
			"namespace", ns, "component", componentName, "inventoryName", objectsInventoryName)
	}
	c.inventoryData.CurrentObjects = GetObjectsFromInventory(objectsInventory)

	// desired
	for _, builder := range resourceBuilders {
		obj, state, err := builder()
		if err != nil {
			return nil, errors.WrapIfWithDetails(err,
				"couldn't build desired object...",
				"namespace", ns, "component", componentName)
		}
		if state != reconciler.StateAbsent {
			desiredObjects = append(desiredObjects, obj)
		}
	}

	// sanitize desired objects
	err = c.sanitizeDesiredObjects(desiredObjects)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err,
			"couldn't sanitize desired objects",
			"namespace", ns, "component", componentName)
	}

	// ensure namespace
	err = c.ensureNamespace(ns, desiredObjects)
	if err != nil {
		return nil, errors.WrapIfWithDetails(err,
			"couldn't ensure namespace meta field on desired objects",
			"namespace", ns, "component", componentName)
	}

	// create inventory of created objects
	if newInventory, err := CreateObjectsInventory(ns, objectsInventoryName, desiredObjects); err == nil {
		c.inventoryData.DesiredObjects = desiredObjects
		return newInventory, nil
	}

	return nil, errors.WrapIfWithDetails(err,
		"during object inventory creation...",
		"namespace", ns, "inventoryName", objectsInventoryName)
}

// Collect `missing` resources from desired state
func (c *Inventory) PrepareDeletableObjects() error {
	var deleteObjects []runtime.Object

	currentObjects := c.inventoryData.CurrentObjects
	for _, currentObject := range currentObjects {
		metaobj, err := meta.Accessor(currentObject)
		if err != nil {
			return errors.WrapIfWithDetails(err,
				"could not access object metadata",
				"gvk", currentObject.GetObjectKind().GroupVersionKind().String())
		}

		isClusterScoped, err := c.IsClusterScoped(currentObject)
		if err != nil {
			c.log.Error(err, "scope check failed, unable to determine whether object is eligible for deletion")
			continue
		}
		// check if current object still exists
		if !isClusterScoped && metaobj.GetNamespace() == "" {
			c.log.Info("object namespace is unknown, unable to determine whether is eligible for deletion", "gvk", currentObject.GetObjectKind().GroupVersionKind().String(), "name", metaobj.GetName())
			continue
		}
		err = c.genericClient.Get(context.TODO(), types.NamespacedName{Namespace: metaobj.GetNamespace(), Name: metaobj.GetName()}, currentObject.(client.Object))
		if err != nil && !meta.IsNoMatchError(err) && !apierrors.IsNotFound(err) {
			return errors.WrapIfWithDetails(err,
				"could not verify if object exists",
				"namespace", metaobj.GetNamespace(), "objectName", metaobj.GetName())
		}

		currentObjGVK := currentObject.GetObjectKind().GroupVersionKind()

		if metaobj.GetDeletionTimestamp() != nil || currentObjGVK.Kind == CustomResourceDefinition || currentObjGVK.Kind == Namespace {
			continue
		}

		desiredObjects := c.inventoryData.DesiredObjects
		found := false

		for _, desiredObject := range desiredObjects {
			desiredObjGVK := desiredObject.GetObjectKind().GroupVersionKind()
			desiredObjMeta, _ := meta.Accessor(desiredObject)

			if currentObjGVK.Group == desiredObjGVK.Group &&
				currentObjGVK.Version == desiredObjGVK.Version &&
				currentObjGVK.Kind == desiredObjGVK.Kind &&
				metaobj.GetNamespace() == desiredObjMeta.GetNamespace() &&
				metaobj.GetName() == desiredObjMeta.GetName() {
				found = true
				break
			}
		}
		if !found {
			c.log.Info("object eligible for delete", "gvk", currentObjGVK.String(), "namespace", metaobj.GetNamespace(), "name", metaobj.GetName())
			deleteObjects = append(deleteObjects, currentObject)
		}
	}

	c.inventoryData.ObjectsToDelete = deleteObjects
	return nil
}

// sanitizeDesiredObjects cleans up the passed desired objects
func (c *Inventory) sanitizeDesiredObjects(desiredObjects []runtime.Object) error {
	for i := range desiredObjects {
		objMeta, err := meta.Accessor(desiredObjects[i])
		if err != nil {
			return errors.WrapIfWithDetails(err, "couldn't get meta data access for object", "gvk", desiredObjects[i].GetObjectKind().GroupVersionKind().String(), "name", objMeta.GetName())
		}

		isClusterScoped, err := c.IsClusterScoped(desiredObjects[i])
		if err != nil {
			c.log.Error(err, "scope check failed")
			continue
		}

		if isClusterScoped && objMeta.GetNamespace() != "" {
			c.log.V(2).Info("removing namespace field from cluster scoped object", "gvk", desiredObjects[i].GetObjectKind().GroupVersionKind().String(), "name", objMeta.GetName())
			objMeta.SetNamespace("")
		}
	}
	return nil
}

// IsClusterScoped returns true of the type if the specified resource is of cluster scope.
// It returns false for namespace scoped resources.
func (c *Inventory) IsClusterScoped(obj runtime.Object) (bool, error) {
	gv, k := obj.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	gvk := strings.Join([]string{gv, k}, "/")

	if c.clusterScopedAPIResources != nil {
		_, ok := c.clusterScopedAPIResources[gvk]
		return ok, nil
	}

	actualGK := obj.GetObjectKind().GroupVersionKind().GroupKind()

	if namespaced, ok := getStaticResourceScope(actualGK); ok {
		return !namespaced, nil
	}

	var fresh bool
	var err error

	fresh, err = initializeAPIResources(c.discoveryClient)
	if err != nil {
		return false, err
	}

	if namespaced, ok := getDynamicResourceScope(actualGK); ok {
		return !namespaced, nil
	}

	if !fresh {
		c.log.Info("API resource not found for object in the cache, updating resource list", "gk", actualGK.String())
		if err := discoverAPIResources(c.discoveryClient); err != nil {
			return false, err
		}
	}

	if namespaced, ok := getDynamicResourceScope(actualGK); ok {
		return !namespaced, nil
	}

	return false, errors.Errorf("unknown resource %s", actualGK.String())
}

// ensureNamespace sets `namespace` as namespace for namespace scoped objects that have no namespace set
func (c *Inventory) ensureNamespace(namespace string, objects []runtime.Object) error {
	for i := range objects {
		objMeta, err := meta.Accessor(objects[i])
		if err != nil {
			return errors.WrapIfWithDetails(err, "couldn't get meta data access for object", "gvk", objects[i].GetObjectKind().GroupVersionKind().String(), "name", objMeta.GetName())
		}

		isClusterScoped, err := c.IsClusterScoped(objects[i])
		if err != nil {
			c.log.Error(err, "scope check failed")
			continue
		}

		if !isClusterScoped && objMeta.GetNamespace() == "" {
			c.log.V(2).Info("setting namespace field for namespace scoped object", "gvk", objects[i].GetObjectKind().GroupVersionKind().String(), "name", objMeta.GetName())
			objMeta.SetNamespace(namespace)
		}
	}
	return nil
}

func (i *Inventory) Append(namespace, component string, parent reconciler.ResourceOwner, resourceBuilders []reconciler.ResourceBuilder) []reconciler.ResourceBuilder {
	if objectInventory, err := i.PrepareDesiredObjects(namespace, component, parent, resourceBuilders); err == nil {
		err := i.PrepareDeletableObjects()
		resourceBuilders = append(resourceBuilders, func() (runtime.Object, reconciler.DesiredState, error) {
			return objectInventory, reconciler.StatePresent, err
		})
	}
	return resourceBuilders
}

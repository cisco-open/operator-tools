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
	"context"
	"fmt"
	"strings"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Backoff = wait.Backoff

type ResourceConditionChecks struct {
	client  client.Client
	backoff wait.Backoff
	log     logr.Logger
	scheme  *runtime.Scheme
}

func NewResourceConditionChecks(client client.Client, backoff Backoff, log logr.Logger, scheme *runtime.Scheme) *ResourceConditionChecks {
	return &ResourceConditionChecks{
		client:  client,
		backoff: wait.Backoff(backoff),
		log:     log,
		scheme:  scheme,
	}
}

func (c *ResourceConditionChecks) WaitForCustomConditionChecks(id string, checkFuncs ...CustomResourceConditionCheck) error {
	log := c.log.WithName(id)

	if l, ok := log.GetSink().(interface{ Grouped(state bool) }); ok {
		l.Grouped(true)
		defer l.Grouped(false)
	}

	log.Info("waiting")

	err := wait.ExponentialBackoff(c.backoff, func() (bool, error) {
		for _, fn := range checkFuncs {
			if ok, err := fn(); !ok {
				if err != nil {
					return false, err
				}
				return false, nil
			}
		}

		return true, nil
	})
	if err != nil {
		return err
	}

	log.Info("done")

	return nil
}

func (c *ResourceConditionChecks) WaitForResources(id string, objects []runtime.Object, checkFuncs ...ResourceConditionCheck) error {
	if len(objects) == 0 || len(checkFuncs) == 0 {
		return nil
	}

	log := c.log.WithName(id)

	if l, ok := log.GetSink().(interface{ Grouped(state bool) }); ok {
		l.Grouped(true)
		defer l.Grouped(false)
	}

	log.Info("waiting")

	for _, o := range objects {
		err := c.waitForResourceConditions(o, log, checkFuncs...)
		if err != nil {
			return err
		}
	}

	log.Info("done")

	return nil
}

func (c *ResourceConditionChecks) waitForResourceConditions(object runtime.Object, log logr.Logger, checkFuncs ...ResourceConditionCheck) error {
	resource := object.DeepCopyObject()

	m, err := meta.Accessor(resource)
	if err != nil {
		return errors.WrapIf(err, "failed to get object key")
	}
	key := client.ObjectKey{Namespace: m.GetNamespace(), Name: m.GetName()}
	log = log.WithValues(c.resourceDetails(resource)...)

	log.V(1).Info("pending")
	err = wait.ExponentialBackoff(c.backoff, func() (bool, error) {
		err := c.client.Get(context.Background(), key, resource.(client.Object))
		for _, fn := range checkFuncs {
			ok := fn(resource, err)
			if !ok {
				if err != nil {
					c.log.V(2).Info("still waiting", "error", err)
				}
				return false, nil
			}
		}

		return true, nil
	})

	if err != nil {
		return err
	}

	log.V(1).Info("ok")

	return nil
}

func (r *ResourceConditionChecks) resourceDetails(desired runtime.Object) (values []interface{}) {
	m, err := meta.Accessor(desired)
	key := client.ObjectKey{Namespace: m.GetNamespace(), Name: m.GetName()}
	if err == nil {
		values = append(values, "name", key.Name)
		if key.Namespace != "" {
			values = append(values, "namespace", key.Namespace)
		}
	}

	gvk := desired.GetObjectKind().GroupVersionKind()
	if gvk.Kind == "" {
		gvko, err := apiutil.GVKForObject(desired, r.scheme)
		if err == nil {
			gvk = gvko
		}
	}

	values = append(values,
		"apiVersion", gvk.GroupVersion().String(),
		"kind", gvk.Kind)

	return
}

func GetFormattedName(name, namespace string, gvk schema.GroupVersionKind) string {
	var group string
	if gvk.Group != "" {
		group = "." + gvk.Group
	}

	if namespace != "" {
		namespace = namespace + "/"
	}
	return fmt.Sprintf("%s%s:%s%s", strings.ToLower(gvk.Kind), group, namespace, name)
}

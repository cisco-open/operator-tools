# Reconciler tools

Reconciling a single resource or a group of resources - describing the desired state declaratively.

## Resource reconciler

`ResourceReconciler` reconciles a single Kubernetes object against the API Server.
It creates the object if it doesn't exist or removes it in case its desired state is absent.

### Static desired states

We can use the following static desired states to make sure a specific resource is:
- created if it doesn't exist
- created and updated if it is out of sync
- removed if it should be absent

```go
const (
  StateCreated                StaticDesiredState = "Created"
  StateAbsent                 StaticDesiredState = "Absent"
  StatePresent                StaticDesiredState = "Present"
)
```

It is using [ObjectMatcher](https://github.com/banzaicloud/k8s-objectmatcher) to be able to tell if an already
existing object needs to be updated or not. It tries hard not to mark resources as changed if the difference
between local and remote states are caused by fields that are not managed locally.

Example:
```go
package main

import (
	corev1 "k8s.io/api/core/v1"
	github.com/go-logr/logr
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func example(client runtimeClient.Client, logger logr.Logger, create bool) {
  resourceReconciler := reconciler.NewReconcilerWith(client, reconciler.WithLog(logger))
  
  serviceObject := &corev1.Service{
    Spec: corev1.ServiceSpec{
      ...
    },
  }
  
  desiredState := reconciler.StateAbsent
  if create {
    desiredState = reconciler.StatePresent
  }
  result, err := resourceReconciler.ReconcileResource(serviceObject, desiredState)
}
```

### Dynamic desired states

Dynamic desired state is an object that holds one or more dynamic functions that will be executed at certain points in a desired object's lifecycle.

Dynamic desired states might be triggered before creating, updating or deleting the object. In addition to that the state of the different objects (desired or current) can be mutated and [runtime client options](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/client#hdr-Options) can be fine tuned as well.

A full implementation of all the available interfaces that the `Resource reconciler` cares about are implemented in the DynamicDesiredState struct.

```
type DynamicDesiredState struct {
	DesiredState     DesiredState
	BeforeCreateFunc func(desired runtime.Object) error
	BeforeUpdateFunc func(current, desired runtime.Object) error
	BeforeDeleteFunc func(current runtime.Object) error
	CreateOptions    []runtimeClient.CreateOption
	UpdateOptions    []runtimeClient.UpdateOption
	DeleteOptions    []runtimeClient.DeleteOption
	ShouldCreateFunc func(desired runtime.Object) (bool, error)
	ShouldUpdateFunc func(current, desired runtime.Object) (bool, error)
	ShouldDeleteFunc func(desired runtime.Object) (bool, error)
}
```

As an example, let's look at the service resource, where any update attempts with an empty `clusterIP` will fail. However, unless the service is headless, the `clusterIP` should only be managed on the server side and will result in errors if we don't set it client side. To overcome this, we can add a `BeforeUpdateFunc` that updates the `clusterIP` field locally to the remote value:

```
ds := &DynamicDesiredState{}
ds.BeforeUpdateFunc = func(current, desired runtime.Object) error {
	if co, ok := current.(*corev1.Service); ok {
		do := desired.(*corev1.Service)
		do.Spec.ClusterIP = co.Spec.ClusterIP
	}
	return nil
}
```

Component reconciler
 - Why this is needed
 - The native reconciler -  example for this in the Thanos Operator
 - The Helm chart reconciler - example for this in the Istio Operator (release-1.11)
 - The inventory

## ResourceReconciler



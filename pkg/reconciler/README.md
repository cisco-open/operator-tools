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

Desired states might be

 - the create/update/delete mutation hooks
 - the create/update/delete validation hooks
 - client options

Component reconciler
 - Why this is needed
 - The native reconciler -  example for this in the Thanos Operator
 - The Helm chart reconciler - example for this in the Istio Operator (release-1.11)
 - The inventory

## ResourceReconciler



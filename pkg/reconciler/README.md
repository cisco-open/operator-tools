
## ResourceReconciler

`ResourceReconciler` reconciles a single Kubernetes object against the API Server.

It creates the object if it doesn't exist or removes it in case its desired state is absent.

It uses the [ObjectMatcher](https://github.com/cisco-open/k8s-objectmatcher) library to be able to tell if an already
existing object needs to be updated or not.

It depends on [logr](github.com/go-logr/logr) logger and the [controller-runtime](sigs.k8s.io/controller-runtime) client
that is available in a typical [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) or [operator-sdk](https://github.com/operator-framework/operator-sdk) project.

Example:
```go
package main

import (
	corev1 "k8s.io/api/core/v1"
	github.com/go-logr/logr
	"github.com/cisco-open/operator-tools/pkg/reconciler"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func example(client runtimeClient.Client, logger logr.Logger) {
  resourceReconciler := reconciler.NewReconcilerWith(client, reconciler.WithLog(logger))
  
  serviceObject := &corev1.Service{
    Spec: corev1.ServiceSpec{
      ...
    },
  }
  
  result, err := resourceReconciler.ReconcileResource(serviceObject, reconciler.StatePresent)
}

```

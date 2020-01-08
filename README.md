# operator-tools

This is a collection of tools to speed up the implementation of Kubernetes Operators with [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder).

## GenericResourceReconciler

`GenericResourceReconciler` reconciles a single Kubernetes object against the API Server.

It creates the object if it doesn't exist or removes it in case its desired state is absent.

It uses the [ObjectMatcher](https://github.com/banzaicloud/k8s-objectmatcher) library to be able to tell if an already
existing object needs to be updated or not.

It depends on [logr](github.com/go-logr/logr) logger and the [controller-runtime](sigs.k8s.io/controller-runtime) client
that is available in a typical [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) project.

Example:
```go
package main

import (
	corev1 "k8s.io/api/core/v1"
	github.com/go-logr/logr
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func example(client runtimeClient.Client, logger logr.Logger) {
	
  resourceReconciler := reconciler.NewReconciler(client, logger, reconciler.ReconcilerOpts{})
  
  serviceObject := &corev1.Service{
    Spec: corev1.ServiceSpec{
      Ports: []corev1.ServicePort{
        {
          Protocol:		corev1.ProtocolTCP,
          Name:				"example",
          Port:				80,
          TargetPort: 8080,
        },
      },
      Selector:	map[string]string{
        "app": "example",
      },
      Type:			corev1.ServiceTypeClusterIP,
      ClusterIP: "None",
    },
  }
  
  result, err := resourceReconciler.ReconcileResource(serviceObject, reconciler.StatePresent)
}

```
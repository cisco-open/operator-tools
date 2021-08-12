
## ResourceReconciler

`ResourceReconciler` reconciles a single Kubernetes object against the API Server.

It creates the object if it doesn't exist or removes it in case its desired state is absent.

It uses the [ObjectMatcher](https://github.com/banzaicloud/k8s-objectmatcher) library to be able to tell if an already
existing object needs to be updated or not.

It depends on [logr](github.com/go-logr/logr) logger and the [controller-runtime](sigs.k8s.io/controller-runtime) client
that is available in a typical [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) or [operator-sdk](https://github.com/operator-framework/operator-sdk) project.

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
  resourceReconciler := reconciler.NewReconcilerWith(client, reconciler.WithLog(logger))
  
  serviceObject := &corev1.Service{
    Spec: corev1.ServiceSpec{
      ...
    },
  }
  
  result, err := resourceReconciler.ReconcileResource(serviceObject, reconciler.StatePresent)
}

```

### Recreating objects on conflict

Some fields of specific objects cannot be updated in certain circumstances. For example the selector for workloads,
like Deployments, StatefulSets or DaemonSets cannot be changed, thus they have to be recreated.

StatefulSets have another set of fields that cannot be changed after creation. If someone want's to change PodManagementPolicy
for example, they have to do that by recreating the StatefulSet. This can be an issue, since it will result in downtime. In
order to overcome this the recommended way is to remove the StatefulSet by orphaning its child objects:
https://kubernetes.io/docs/tasks/run-application/delete-stateful-set/#deleting-a-statefulset

Note: this wouldn't work when changing the selector, since the new StatefulSet won't be able to adapt child objects.

The `ResourceReconciler` does these two things basically by looking at the response of an update, and if that was a 422 HTTP error code
then it looks for errors that could be fixed by recreating the object.

The legacy default behaviour for recreating objects in foreground when their selector has been changed is now extended with recreating
a StatefulSet using the Orphan policy when there is a change in a restricted field, to allow their pods keep on running and get
adopted by the newly created StatefulSet resource.

Enable recreating objects using the legacy behaviour
```
resourceReconciler := reconciler.NewReconcilerWith(client,
    reconciler.WithEnableRecreateWorkload()
)
```

Enable recreating all objects when there is an error with the update
```
resourceReconciler := reconciler.NewReconcilerWith(client,
    reconciler.WithEnableRecreateWorkload(),
    reconciler.WithRecreateEnabledForAll()
)
```

Enable recreating by a custom condition
```
resourceReconciler := reconciler.NewReconcilerWith(client,
    reconciler.WithEnableRecreateWorkload(),
    WithRecreateEnabledFor(func(kind schema.GroupVersionKind, status metav1.Status) (RecreateConfig, error) {
        match, err := regexp.Match(`field is immutable`, []byte(status.Message))
        if err != nil {
            return RecreateConfig{}, err
        }
        if match {
            return RecreateConfig{
                Delete:              true,
                RecreateImmediately: false,
                DeletePropagation:   metav1.DeletePropagationForeground,
                Delay:               DefaultRecreateRequeueDelay,
            }, nil
        }
        return RecreateConfig{}, nil
    })
)
```
# operator-tools

This is a collection of tools to speed up the implementation of Kubernetes Operators on top of [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime).

## GenericResourceReconciler

`GenericResourceReconciler` reconciles a single Kubernetes object against the API Server.

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

This feature is currently only covered with tests in the [logging operator](https://github.com/banzaicloud/logging-operator),
but this is a subject to change soon.

## Secret & SecretLoader

`Secret` is a type to be used in CRDs to abstract the concept of loading a secret item instead of defining it with it's value directly.

Currently it support Kubernetes secrets only, but it can be extended to refer to secrets in custom secret stores as well.

There are two main approaches to load secrets and one for testing. 
 
1. Load the secrets and return with their value directly if `ValueFrom` is set.
1. Load the secrets in the background if `MountFrom` is set, but return only the full path where they should be available in a container. 
It's the callers responsibility to make those secrets available on that given path, e.g. by creating an aggregated secret with all
the referenced secrets and mount it into the container through a secret volume (this is how we use it).
1. Load the value directly if `Value` is set. (This is only good for testing.)

Once you're done with configuration you can create the `SecretLoader` and load your secrets through it.

```go
mountSecrets := &secret.MountSecrets{}
secretLoader := secret.NewSecretLoader(client, namespace, "/path/to/mount", mountSecrets)
```

Then you can load the secrets. The following steps can be made more dynamic, like it is beeing used in the logging operator:
https://github.com/banzaicloud/logging-operator/blob/master/pkg/sdk/model/types/stringmaps.go

```go
// get the secretValue and use it as you like in an application configuration template for example
secretValue, err := secretLoader.Load(yourCustomResourceType.Spec.ExampleSecretField)

// get the path to the mount secret and use it as you like in an application configuration template for example
secretPath, err := secretLoader.Load(yourCustomResourceType.Spec.ExampleMountSecretField)

// render the configuration template and create a new secret from it that will be mounted into the container
appConfigSecret := &corev1.Secret{}
renderedTemplate := renderTemplate(secretValue, secretPath)
appConfigSecret.Data["app.conf"] = renderedTemplate

// create the combined secret to be mounted to the container on "/path/to/mount"
combinedSecret := &corev1.Secret{}
for _, secret := range *mountSecrets {
  combinedSecret.Data[secret.MappedKey] = secret.Value
}
```

For a full example please check out the [logging operator code](https://github.com/banzaicloud/logging-operator).

Also, this feature is currently only covered with tests in the [logging operator](https://github.com/banzaicloud/logging-operator),
but this is a subject to change soon.

## KubernetesVolume

Configure volumes in custom types for underlying pods in a uniform way.

```go
type SomeCustomApp struct {
	BufferStorage *volume.KubernetesVolume `json:"bufferStorage,omitempty"` 
}
```

And this is how you would configure it through yaml
```yaml
someCustomApp:
  bufferStorage:
    hostPath:
      path: "/buffer" # set this here, or provide a default in the code
```

In the operator you can use the `GetVolume` method on the object in a generic way:
```go
// provide a default path if hostPath is used but no actual path has been configured explicitly
someCustomApp.bufferStorage.WithDefaultHostPath("/opt/buffer")

volume := someCustomApp.bufferStorage.GetVolume(
  "volumeName",
))
```
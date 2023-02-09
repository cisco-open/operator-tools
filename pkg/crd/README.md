
## CRD

Once an operator is installed and CRDs are applied, these commands help to verify CRDs are in place.

```go
package main

import (
    "flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/cisco-open/operator-tools/pkg/crd"
)

func main() {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	c := crd.NewCRD(clientset)
	resources, err := c.ListAPIResources(metav1.GroupVersion{
		Group:   "monitoring.coreos.com",
		Version: "v1",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, resource := range resources {
		fmt.Println(resource.Name)
	}
	// Output:
	// alertmanagers
	// servicemonitors
	// prometheuses
	// prometheusrules
	// podmonitors

	has, err := c.HasAPIResource(metav1.GroupVersion{
		Group:   "monitoring.coreos.com",
		Version: "v1",
	}, "servicemonitors")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%t\n", has)
	// Output: true
}


func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
```

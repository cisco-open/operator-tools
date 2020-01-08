module github.com/banzaicloud/operator-tools

go 1.13

require (
	emperror.dev/errors v0.4.2
	github.com/banzaicloud/k8s-objectmatcher v1.0.1
	github.com/go-logr/logr v0.1.0
	github.com/goph/emperror v0.17.2
	github.com/onsi/gomega v1.5.0
	github.com/pborman/uuid v1.2.0
	github.com/spf13/cast v1.3.0
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	k8s.io/api v0.16.4
	k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery v0.16.4
	k8s.io/client-go v11.0.1-0.20190516230509-ae8359b20417+incompatible
	sigs.k8s.io/controller-runtime v0.4.0
)

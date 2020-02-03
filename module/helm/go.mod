module github.com/banzaicloud/operator-tools/module/helm

go 1.13

require (
	emperror.dev/errors v0.7.0
	github.com/go-logr/logr v0.1.0
	helm.sh/helm/v3 v3.0.3
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	sigs.k8s.io/controller-runtime v0.4.0
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

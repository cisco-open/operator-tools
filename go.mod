module github.com/banzaicloud/operator-tools

go 1.13

require (
	emperror.dev/errors v0.8.0
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/banzaicloud/k8s-objectmatcher v1.5.0
	github.com/briandowns/spinner v1.11.1
	github.com/cppforlife/go-patch v0.2.0
	github.com/fatih/color v1.7.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.2.1
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/iancoleman/orderedmap v0.0.0-20190318233801-ac98e3ecb4b0
	github.com/pborman/uuid v1.2.0
	github.com/spf13/cast v1.3.1
	github.com/stretchr/testify v1.6.1
	github.com/wayneashleyberry/terminal-dimensions v1.0.0
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.3.4
	k8s.io/api v0.19.2
	k8s.io/apiextensions-apiserver v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/utils v0.0.0-20200729134348-d5654de09c73
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/yaml v1.2.0
)

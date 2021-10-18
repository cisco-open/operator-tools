LICENSEI_VERSION = 0.2.0
GOLANGCI_VERSION = 1.21.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CONTROLLER_GEN_VERSION = v0.4.1
CONTROLLER_GEN = $(PWD)/bin/controller-gen

OS = $(shell uname | tr A-Z a-z)

KUBEBUILDER_VERSION = 2.3.1
export KUBEBUILDER_ASSETS := $(PWD)/bin

# Generate code
generate: bin/controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/secret/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/volume/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/prometheus/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/types/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/typeoverride/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/helm/...

bin/controller-gen:
	@ if ! test -x bin/controller-gen; then \
		set -ex ;\
		CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
		cd $$CONTROLLER_GEN_TMP_DIR ;\
		go mod init tmp ;\
		GOBIN=$(PWD)/bin go get sigs.k8s.io/controller-tools/cmd/controller-gen@${CONTROLLER_GEN_VERSION} ;\
		rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	fi

.PHONY: bin/kubebuilder_$(KUBEBUILDER_VERSION)
bin/kubebuilder_$(KUBEBUILDER_VERSION):
	@if ! test -L bin/kubebuilder_$(KUBEBUILDER_VERSION); \
		then \
		mkdir -p bin; \
		curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VERSION)/kubebuilder_$(KUBEBUILDER_VERSION)_$(OS)_amd64.tar.gz | tar xvz -C bin; \
		ln -sf kubebuilder_$(KUBEBUILDER_VERSION)_$(OS)_amd64/bin bin/kubebuilder_$(KUBEBUILDER_VERSION); \
	fi

bin/kubebuilder: bin/kubebuilder_$(KUBEBUILDER_VERSION)
	@ln -sf kubebuilder_$(KUBEBUILDER_VERSION)/kubebuilder bin/kubebuilder
	@ln -sf kubebuilder_$(KUBEBUILDER_VERSION)/kube-apiserver bin/kube-apiserver
	@ln -sf kubebuilder_$(KUBEBUILDER_VERSION)/etcd bin/etcd
	@ln -sf kubebuilder_$(KUBEBUILDER_VERSION)/kubectl bin/kubectl

bin/licensei: bin/licensei-${LICENSEI_VERSION}
	@ln -sf licensei-${LICENSEI_VERSION} bin/licensei
bin/licensei-${LICENSEI_VERSION}:
	@mkdir -p bin
	curl -sfL https://git.io/licensei | bash -s v${LICENSEI_VERSION}
	@mv bin/licensei $@

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	bin/licensei cache

.PHONY: license-check
license-check: bin/licensei ## Run license check
	bin/licensei check
	bin/licensei header

.PHONY: test
test: bin/kubebuilder
	go test ./...

.PHONY: check
check: test lint check-diff ## Run tests and linters

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ./bin/ v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: export CGO_ENABLED = 1
lint: bin/golangci-lint ## Run linter
	bin/golangci-lint run

.PHONY: fix
fix: export CGO_ENABLED = 1
fix: bin/golangci-lint ## Fix lint violations
	bin/golangci-lint run --fix

check-diff: generate-type-docs
	go mod tidy
	$(MAKE) generate docs
	git diff --exit-code

generate-type-docs:
	go run cmd/docs.go

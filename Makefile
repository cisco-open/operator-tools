####
##  Dependency versions
####

CONTROLLER_TOOLS_VERSION := 0.21.0

GOLANGCI_LINT_VERSION := 2.12.2

LICENSEI_VERSION := 0.9.0

ENVTEST_K8S_VERSION := 1.35.0

BIN := ${PWD}/bin

export PATH := $(BIN):$(PATH)

GOVERSION := $(shell go env GOVERSION)

GOLANGCI_LINT  := $(BIN)/golangci-lint
CONTROLLER_GEN ?= $(BIN)/controller-gen
ENVTEST        ?= $(BIN)/setup-envtest
LICENSEI       := $(BIN)/licensei

ENVTEST_BIN_DIR       := $(BIN)/envtest
ENVTEST_BINARY_ASSETS := $(ENVTEST_BIN_DIR)/bin

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

.DEFAULT_GOAL = help
.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: generate
generate: codegen docs fmt ## Generate code, documentation, etc.

.PHONY: codegen
codegen: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/secret/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/volume/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/prometheus/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/types/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/typeoverride/...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/helm/...

.PHONY: docs
docs: ## Generate type documentation.
	go run cmd/docs.go

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: tidy
tidy: ## Tidy Go modules.
	find . -iname "go.mod" -not -path "./.devcontainer/*" | xargs -L1 sh -c 'cd $$(dirname $$0); go mod tidy'

.PHONY: test
test: fmt vet envtest ## Run verifications and tests.
	KUBEBUILDER_ASSETS="$(ENVTEST_BINARY_ASSETS)" go test -v ./... -coverprofile cover.out

.PHONY: lint
lint: export CGO_ENABLED = 1
lint: ${GOLANGCI_LINT} ## Run golangci-lint.
	${GOLANGCI_LINT} run ${LINTER_FLAGS}

.PHONY: lint-fix
lint-fix: export CGO_ENABLED = 1
lint-fix: ${GOLANGCI_LINT} ## Run golangci-lint and perform fixes.
	${GOLANGCI_LINT} run --fix

.PHONY: check
check: test lint check-diff ## Run tests and linters.

.PHONY: check-diff
check-diff: tidy generate ## Verify that generated files are up to date.
	git diff --exit-code

.PHONY: license-cache
license-cache: ${LICENSEI} ## Generate license cache.
	${LICENSEI} cache

.PHONY: license-check
license-check: ${LICENSEI} .licensei.cache ## Run license check.
	${LICENSEI} check
	${LICENSEI} header

##@ Build Dependencies

${GOLANGCI_LINT}: ${GOLANGCI_LINT}_${GOLANGCI_LINT_VERSION}_${GOVERSION} | ${BIN}
	ln -snf $(notdir $<) $@

${GOLANGCI_LINT}_${GOLANGCI_LINT_VERSION}_${GOVERSION}: IMPORT_PATH := github.com/golangci/golangci-lint/v2/cmd/golangci-lint
${GOLANGCI_LINT}_${GOLANGCI_LINT_VERSION}_${GOVERSION}: VERSION := v${GOLANGCI_LINT_VERSION}
${GOLANGCI_LINT}_${GOLANGCI_LINT_VERSION}_${GOVERSION}: | ${BIN}
	${go_install_binary}

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): | $(BIN)
	test -s $(BIN)/controller-gen && $(BIN)/controller-gen --version | grep -q v$(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(BIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST_BINARY_ASSETS) ## Download envtest-setup and Kubernetes binary assets locally if necessary.
$(ENVTEST): | $(BIN)
	test -s $(BIN)/setup-envtest || GOBIN=$(BIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

$(ENVTEST_BINARY_ASSETS): $(ENVTEST_BINARY_ASSETS)_$(ENVTEST_K8S_VERSION)
	ln -snf $(notdir $<) $@

$(ENVTEST_BINARY_ASSETS)_$(ENVTEST_K8S_VERSION): | $(ENVTEST) $(ENVTEST_BIN_DIR)
	ln -snf $$($(ENVTEST) --bin-dir $(ENVTEST_BIN_DIR) use $(ENVTEST_K8S_VERSION) -p path) $@

$(ENVTEST_BIN_DIR): | $(BIN)
	mkdir -p $@

${LICENSEI}: ${LICENSEI}_${LICENSEI_VERSION}_${GOVERSION} | ${BIN}
	ln -snf $(notdir $<) $@

${LICENSEI}_${LICENSEI_VERSION}_${GOVERSION}: IMPORT_PATH := github.com/goph/licensei/cmd/licensei
${LICENSEI}_${LICENSEI_VERSION}_${GOVERSION}: VERSION := v${LICENSEI_VERSION}
${LICENSEI}_${LICENSEI_VERSION}_${GOVERSION}: | ${BIN}
	${go_install_binary}

.licensei.cache: ${LICENSEI}
ifndef GITHUB_TOKEN
	@>&2 echo "WARNING: building licensei cache without Github token, rate limiting might occur."
	@>&2 echo "(Hint: If too many licenses are missing, try specifying a Github token via the environment variable GITHUB_TOKEN.)"
endif
	${LICENSEI} cache

${BIN}:
	mkdir -p $@

define go_install_binary
find ${BIN} -name '$(notdir ${IMPORT_PATH})_*' -exec rm {} +
GOBIN=${BIN} go install ${IMPORT_PATH}@${VERSION}
mv ${BIN}/$(notdir ${IMPORT_PATH}) $@
endef

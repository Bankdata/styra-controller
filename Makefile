# Copyright (C) 2023 Bankdata (bankdata@bankdata.dk)

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# GO builder image overwrite to the Dockerfile
BUILDER_IMAGE ?= ""

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26.0

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

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: generate-mocks
generate-mocks: mockery ## Generate mocks
	$(MOCKERY)

.PHONY: generate-docs
generate-docs: gen-crd-api-reference-docs ## Generate API documentation based on the API types.
	./scripts/gen-api-docs/gen-api-docs.sh all

.PHONY: lint
lint: golangci-lint ## Run linters
	$(GOLANGCI_LINT) run

.PHONY: test
test: ginkgo manifests generate lint envtest generate-mocks ## Run all tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" $(GINKGO) -r --output-interceptor-mode=none

.PHONY: test-unit
test-unit: ginkgo manifests generate lint generate-mocks ## Run unit tests.
	$(GINKGO) -r --label-filter "!integration" --output-interceptor-mode=none

.PHONY: test-integration ## Run integration tests.
test-integration: ginkgo manifests generate lint envtest generate-mocks ## Run integration tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" $(GINKGO) -r --label-filter "integration" --output-interceptor-mode=none

.PHONY: kind-create
kind-create: kind ## Create kind cluster
	$(KIND) get clusters | grep styra-controller || \
	$(KIND) create cluster --name styra-controller && \
	kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml

.PHONY: kind-delete
kind-delete: kind ## Delete kind cluster
	$(KIND) delete cluster --name styra-controller

.PHONY: kind-load
kind-load: docker-build kind ## Build and load docker image in kind
	$(KIND) load docker-image ${IMG} --name styra-controller

##@ Build

.PHONY: build
build: manifests generate goreleaser ## Build manager binary.
	$(GORELEASER) build --single-target --snapshot --clean

.PHONY: run
run: manifests generate ## Run a controller from your host.
	go run ./cmd/main.go

.PHONY: docker-build
docker-build: goreleaser ## Build docker image with the manager.
	GOOS=linux $(GORELEASER) build \
	  --single-target --snapshot --clean \
	  -o dist/styra-controller
	docker build -t ${IMG} -f build/package/Dockerfile \
	  --build-arg BINARY="dist/styra-controller" .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

.PHONY: release
release: goreleaser ## Release project
	$(GORELEASER) release

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
KUSTOMIZE_GO_MOD_VERSION ?= $(shell cat go.mod | grep -E "sigs.k8s.io/kustomize/kustomize/v5" | cut -d ' ' -f 2)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_GEN_GO_MOD_VERSION ?= $(shell cat go.mod | grep -E "sigs.k8s.io/controller-tools" | cut -d ' ' -f 2)

ENVTEST ?= $(LOCALBIN)/setup-envtest

GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
GOLANGCI_LINT_GO_MOD_VERSION ?= $(shell cat go.mod | grep -E "github.com/golangci/golangci-lint" | cut -d ' ' -f 2)

GINKGO ?= $(LOCALBIN)/ginkgo
GINKGO_GO_MOD_VERSION ?= $(shell cat go.mod | grep -E "github.com/onsi/ginkgo/v2" | cut -d ' ' -f2 | cut -c 2-)

MOCKERY ?= $(LOCALBIN)/mockery
MOCKERY_GO_MOD_VERSION ?= $(shell cat go.mod | grep -E "github.com/vektra/mockery/v2" | cut -d ' ' -f 2)

KIND ?= $(LOCALBIN)/kind
KIND_GO_MOD_VERSION ?= $(shell cat go.mod | grep -E "sigs.k8s.io/kind" | cut -d ' ' -f 2)

GEN_CRD_API_REFERENCE_DOCS ?= $(LOCALBIN)/gen-crd-api-reference-docs

GORELEASER ?= $(LOCALBIN)/goreleaser
GORELEASER_GO_MOD_VERSION ?= $(shell cat go.mod | grep -E "github.com/goreleaser/goreleaser" | cut -d ' ' -f 2)

.PHONY: kustomize
kustomize: ## Download kustomize locally if necessary.
	test -s $(KUSTOMIZE) && \
	$(KUSTOMIZE) version | grep $(KUSTOMIZE_GO_MOD_VERSION) || \
	GOBIN=$(LOCALBIN) go install -ldflags "-X sigs.k8s.io/kustomize/api/provenance.version=$(KUSTOMIZE_GO_MOD_VERSION)" sigs.k8s.io/kustomize/kustomize/v5

.PHONY: controller-gen
controller-gen: gopls ## Download controller-gen locally if necessary.
	test -s $(CONTROLLER_GEN) && \
	$(CONTROLLER_GEN) --version | grep $(CONTROLLER_GEN_GO_MOD_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen

.PHONY: gopls
gopls: 
	# $(shell go version)
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/gopls@latest


.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(ENVTEST) || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest

.PHONY: golangci-lint
golangci-lint: ## Download golangci-lint locally if necessary.
	test -s $(GOLANGCI_LINT) && \
	$(GOLANGCI_LINT) --version | grep -o $(GOLANGCI_LINT_GO_MOD_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: ginkgo
ginkgo: ## Download ginkgo locally if necessary.
	test -s $(GINKGO) && \
	$(GINKGO) version | grep -o $(GINKGO_GO_MOD_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo

.PHONY: mockery 
mockery: ## Download mockery locally if necessary.
	test -s $(MOCKERY) && \
	$(MOCKERY) --version | grep $(MOCKERY_GO_MOD_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/vektra/mockery/v2

.PHONY: kind 
kind: ## Download kind locally if necessary.
	test -s $(KIND) && \
	$(KIND) version | grep $(KIND_GO_MOD_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kind

.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: ## Download gen-crd-api-reference-docs locally if necessary.
	test -s $(GEN_CRD_API_REFERENCE_DOCS) || \
	GOBIN=$(LOCALBIN) go install github.com/ahmetb/gen-crd-api-reference-docs

.PHONY: goreleaser 
goreleaser: ## Download goreleaser locally if necessary.
	test -s $(GORELEASER) && \
	$(GORELEASER) -v | grep $(GORELEASER_GO_MOD_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/goreleaser/goreleaser

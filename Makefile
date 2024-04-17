#!/usr/bin/env bash

#  Copyright 2024 The Kubernetes Authors.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

## --------------------------------------
## General
## --------------------------------------

SHELL:=/usr/bin/env bash
.DEFAULT_GOAL:=help

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://proxy.golang.org
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on

# Tools.
ENVTEST_DIR := hack/envtest
ENVTEST_MATRIX_DIR := $(ENVTEST_DIR)/_matrix
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/bin)
GOLANGCI_LINT := $(abspath $(TOOLS_BIN_DIR)/golangci-lint)
GO_INSTALL := ./hack/go-install.sh

## --------------------------------------
## Binaries
## --------------------------------------

GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT_VER := $(shell cat .github/workflows/golangci-lint.yml | grep [[:space:]]version: | sed 's/.*version: //')
GOLANGCI_LINT := $(abspath $(TOOLS_BIN_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER))
GOLANGCI_LINT_PKG := github.com/golangci/golangci-lint/cmd/golangci-lint

$(GOLANGCI_LINT): # Build golangci-lint from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(GOLANGCI_LINT_PKG) $(GOLANGCI_LINT_BIN) $(GOLANGCI_LINT_VER)

## --------------------------------------
## Linting
## --------------------------------------

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint codebase
	$(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)
	cd tools/setup-envtest; $(GOLANGCI_LINT) run -v $(GOLANGCI_LINT_EXTRA_ARGS)

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT) ## Lint the codebase and run auto-fixers if supported by the linter.
	GOLANGCI_LINT_EXTRA_ARGS=--fix $(MAKE) lint

## --------------------------------------
## Testing
## --------------------------------------

.PHONY: test
test: ## Run the test.sh script which will check all.
	TRACE=1 ./test.sh

test-all:
	$(MAKE) test

.PHONY: modules
modules: ## Runs go mod to ensure modules are up to date.
	go mod tidy

## --------------------------------------
## Cleanup / Verification
## --------------------------------------

.PHONY: clean
clean: ## Cleanup.
	$(GOLANGCI_LINT) cache clean
	$(MAKE) clean-bin

.PHONY: clean-bin
clean-bin: ## Remove all generated binaries.
	rm -rf hack/tools/bin

.PHONE: clean-release
clean-release: ## Remove all generated release binaries.
	rm -rf $(RELEASE_DIR)

## --------------------------------------
## Envtest Build
## --------------------------------------

RELEASE_DIR := out

.PHONY: $(RELEASE_DIR)
$(RELEASE_DIR):
	mkdir -p $(RELEASE_DIR)/

.PHONY: release-envtest
release-envtest: clean-release ## Build the envtest binaries by operating system.
	OS=linux ARCH=amd64 $(MAKE) release-envtest-docker-build
	OS=linux ARCH=arm64 $(MAKE) release-envtest-docker-build
	OS=linux ARCH=ppc64le $(MAKE) release-envtest-docker-build
	OS=linux ARCH=s390x $(MAKE) release-envtest-docker-build
	OS=darwin ARCH=amd64 $(MAKE) release-envtest-docker-build
	OS=darwin ARCH=arm64 $(MAKE) release-envtest-docker-build
	OS=windows ARCH=amd64 $(MAKE) release-envtest-docker-build
	./hack/envtest/update-releases.sh

.PHONY: release-envtest-docker-build
release-envtest-docker-build: $(RELEASE_DIR) ## Build the envtest binaries.
	@: $(if $(KUBERNETES_VERSION),,$(error KUBERNETES_VERSION is not set))
	@: $(if $(OS),,$(error OS is not set))
	@: $(if $(ARCH),,$(error ARCH is not set))
	docker buildx build \
		--file ./hack/envtest/$(OS)/Dockerfile \
		--build-arg KUBERNETES_VERSION=$(KUBERNETES_VERSION) \
		--build-arg GO_VERSION=$(shell yq eval '.go' $(ENVTEST_MATRIX_DIR)/$(KUBERNETES_VERSION).yaml) \
		--build-arg ETCD_VERSION=$(shell yq eval '.etcd' $(ENVTEST_MATRIX_DIR)/$(KUBERNETES_VERSION).yaml) \
		--build-arg OS=$(OS) \
		--build-arg ARCH=$(ARCH) \
		--tag sigs.k8s.io/controller-tools/envtest:$(KUBERNETES_VERSION)-$(OS)-$(ARCH) \
		--output type=local,dest=$(RELEASE_DIR) \
		.

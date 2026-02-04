SHELL=/usr/bin/env bash -o pipefail

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
ifeq ($(GOARCH),arm)
	ARCH=armv7
else
	ARCH=$(GOARCH)
endif
GODEBUG :=

CONTAINER_CLI ?= docker

GO_PKG=github.com/prometheus-operator/prometheus-operator
IMAGE_OPERATOR?=quay.io/prometheus-operator/prometheus-operator
IMAGE_RELOADER?=quay.io/prometheus-operator/prometheus-config-reloader
IMAGE_WEBHOOK?=quay.io/prometheus-operator/admission-webhook
TAG?=$(shell git rev-parse --short HEAD)
VERSION?=$(shell cat VERSION | tr -d " \t\n\r")
GO_VERSION?=$(shell grep golang-version .github/env | sed "s/golang-version=//")

CRD_OPTIONS ?= "crd:crdVersions=v1"

KIND_CONTEXT ?= e2e

TYPES_V1_TARGET := pkg/apis/monitoring/v1/types.go
TYPES_V1_TARGET += pkg/apis/monitoring/v1/alertmanager_types.go
TYPES_V1_TARGET += pkg/apis/monitoring/v1/podmonitor_types.go
TYPES_V1_TARGET += pkg/apis/monitoring/v1/probe_types.go
TYPES_V1_TARGET += pkg/apis/monitoring/v1/prometheus_types.go
TYPES_V1_TARGET += pkg/apis/monitoring/v1/servicemonitor_types.go
TYPES_V1_TARGET += pkg/apis/monitoring/v1/thanos_types.go

TYPES_V1ALPHA1_TARGET := pkg/apis/monitoring/v1alpha1/alertmanager_config_types.go
TYPES_V1ALPHA1_TARGET += pkg/apis/monitoring/v1alpha1/prometheusagent_types.go
TYPES_V1ALPHA1_TARGET += pkg/apis/monitoring/v1alpha1/scrapeconfig_types.go
TYPES_V1BETA1_TARGET := pkg/apis/monitoring/v1beta1/alertmanager_config_types.go

ROOT_DIR=$(shell pwd)
TOOLS_BIN_DIR ?= $(ROOT_DIR)/tmp/bin
export PATH := $(TOOLS_BIN_DIR):$(PATH)

CONTROLLER_GEN_BINARY := $(TOOLS_BIN_DIR)/controller-gen
JB_BINARY=$(TOOLS_BIN_DIR)/jb
GOJSONTOYAML_BINARY=$(TOOLS_BIN_DIR)/gojsontoyaml
JSONNET_BINARY=$(TOOLS_BIN_DIR)/jsonnet
JSONNETFMT_BINARY=$(TOOLS_BIN_DIR)/jsonnetfmt
SHELLCHECK_BINARY=$(TOOLS_BIN_DIR)/shellcheck
PROMLINTER_BINARY=$(TOOLS_BIN_DIR)/promlinter
PROMTOOL_BINARY=$(TOOLS_BIN_DIR)/promtool
GOLANGCILINTER_BINARY=$(TOOLS_BIN_DIR)/golangci-lint
MDOX_BINARY=$(TOOLS_BIN_DIR)/mdox
API_DOC_GEN_BINARY=$(TOOLS_BIN_DIR)/gen-crd-api-reference-docs
GOLANGCIKUBEAPILINTER_BINARY=$(TOOLS_BIN_DIR)/golangci-kube-api-linter
TOOLING=$(CONTROLLER_GEN_BINARY) $(GOBINDATA_BINARY) $(JB_BINARY) $(GOJSONTOYAML_BINARY) $(JSONNET_BINARY) $(JSONNETFMT_BINARY) $(SHELLCHECK_BINARY) $(PROMLINTER_BINARY) $(PROMTOOL_BINARY) $(GOLANGCILINTER_BINARY) $(MDOX_BINARY) $(API_DOC_GEN_BINARY) $(GOLANGCIKUBEAPILINTER_BINARY)

K8S_GEN_BINARIES:=informer-gen lister-gen client-gen applyconfiguration-gen
K8S_GEN_ARGS:=--go-header-file $(shell pwd)/.header --v=1 --logtostderr

K8S_GEN_DEPS:=.header
K8S_GEN_DEPS+=$(TYPES_V1_TARGET)
K8S_GEN_DEPS+=$(TYPES_V1ALPHA1_TARGET)
K8S_GEN_DEPS+=$(TYPES_V1BETA1_TARGET)
K8S_GEN_DEPS+=$(foreach bin,$(K8S_GEN_BINARIES),$(TOOLS_BIN_DIR)/$(bin))

CERTS_DIR := test/e2e/tls_certs

BUILD_DATE=$(shell date +"%Y%m%d-%T")
# source: https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables#default-environment-variables
ifndef GITHUB_ACTIONS
	BUILD_USER?=$(USER)
	BUILD_BRANCH?=$(shell git branch --show-current)
	BUILD_REVISION?=$(shell git rev-parse --short HEAD)
else
	BUILD_USER=Action-Run-ID-$(GITHUB_RUN_ID)
	BUILD_BRANCH=$(GITHUB_REF:refs/heads/%=%)
	BUILD_REVISION=$(GITHUB_SHA)
endif
GITHUB_TOKEN?=

# The Prometheus common library import path
PROMETHEUS_COMMON_PKG=github.com/prometheus/common

# The ldflags for the go build process to set the version related data.
GO_BUILD_LDFLAGS=\
	-s \
	-X $(PROMETHEUS_COMMON_PKG)/version.Revision=$(BUILD_REVISION)  \
	-X $(PROMETHEUS_COMMON_PKG)/version.BuildUser=$(BUILD_USER) \
	-X $(PROMETHEUS_COMMON_PKG)/version.BuildDate=$(BUILD_DATE) \
	-X $(PROMETHEUS_COMMON_PKG)/version.Branch=$(BUILD_BRANCH) \
	-X $(PROMETHEUS_COMMON_PKG)/version.Version=$(VERSION)

GO_BUILD_RECIPE=\
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_ENABLED=0 \
	go build -ldflags="$(GO_BUILD_LDFLAGS)"

pkgs = $(shell go list ./... | grep -v /test/ | grep -v /contrib/)
pkgs += $(shell go list $(GO_PKG)/pkg/apis/monitoring...)
pkgs += $(shell go list $(GO_PKG)/pkg/client...)

.PHONY: all
all: format generate build test

.PHONY: clean
clean:
	# Remove all files and directories ignored by git.
	git clean -Xfd .

############
# Building #
############

.PHONY: build
build: operator prometheus-config-reloader admission-webhook k8s-gen

.PHONY: operator
operator:
	$(GO_BUILD_RECIPE) -o $@ ./cmd/operator/

.PHONY: prometheus-config-reloader
prometheus-config-reloader:
	$(GO_BUILD_RECIPE) -o $@ ./cmd/$@/

.PHONY: admission-webhook
admission-webhook:
	$(GO_BUILD_RECIPE) -o $@ ./cmd/$@/


DEEPCOPY_TARGETS := pkg/apis/monitoring/v1/zz_generated.deepcopy.go pkg/apis/monitoring/v1alpha1/zz_generated.deepcopy.go pkg/apis/monitoring/v1beta1/zz_generated.deepcopy.go
$(DEEPCOPY_TARGETS): $(CONTROLLER_GEN_BINARY)
	cd ./pkg/apis/monitoring/v1 && $(CONTROLLER_GEN_BINARY) object:headerFile=$(CURDIR)/.header \
		paths=.
	cd ./pkg/apis/monitoring/v1alpha1 && $(CONTROLLER_GEN_BINARY) object:headerFile=$(CURDIR)/.header \
		paths=.
	cd ./pkg/apis/monitoring/v1beta1 && $(CONTROLLER_GEN_BINARY) object:headerFile=$(CURDIR)/.header \
		paths=.

.PHONY: k8s-client-gen
k8s-client-gen: $(K8S_GEN_DEPS)
	rm -rf pkg/client/{versioned,informers,listers,applyconfiguration}

	@echo ">> generating pkg/client/applyconfiguration..."
	GODEBUG=$(GODEBUG) $(APPLYCONFIGURATION_GEN_BINARY) \
		$(K8S_GEN_ARGS) \
		--output-pkg "$(GO_PKG)/pkg/client/applyconfiguration" \
		--output-dir "pkg/client/applyconfiguration" \
		"$(GO_PKG)/pkg/apis/monitoring/v1" "$(GO_PKG)/pkg/apis/monitoring/v1alpha1" "$(GO_PKG)/pkg/apis/monitoring/v1beta1"

	@echo ">> generating pkg/client/versioned..."
	GODEBUG=$(GODEBUG) $(CLIENT_GEN_BINARY) \
		$(K8S_GEN_ARGS) \
		--apply-configuration-package "$(GO_PKG)/pkg/client/applyconfiguration" \
		--input-base                  "$(GO_PKG)/pkg/apis" \
		--clientset-name              "versioned" \
		--output-pkg                  "$(GO_PKG)/pkg/client" \
		--output-dir                  "pkg/client" \
		--input monitoring/v1 \
		--input monitoring/v1beta1 \
		--input monitoring/v1alpha1

	@echo ">> generating pkg/client/listers..."
	GODEBUG=$(GODEBUG) $(LISTER_GEN_BINARY) \
		$(K8S_GEN_ARGS) \
		--output-pkg "$(GO_PKG)/pkg/client/listers" \
		--output-dir "pkg/client/listers" \
		"$(GO_PKG)/pkg/apis/monitoring/v1" "$(GO_PKG)/pkg/apis/monitoring/v1alpha1" "$(GO_PKG)/pkg/apis/monitoring/v1beta1"

	@echo ">> generating pkg/client/informers..."
	GODEBUG=$(GODEBUG) $(INFORMER_GEN_BINARY) \
		$(K8S_GEN_ARGS) \
		--versioned-clientset-package "$(GO_PKG)/pkg/client/versioned" \
		--listers-package             "$(GO_PKG)/pkg/client/listers" \
		--output-pkg                  "$(GO_PKG)/pkg/client/informers" \
		--output-dir                  "pkg/client/informers" \
		"$(GO_PKG)/pkg/apis/monitoring/v1" "$(GO_PKG)/pkg/apis/monitoring/v1alpha1" "$(GO_PKG)/pkg/apis/monitoring/v1beta1"

.PHONY: k8s-gen
k8s-gen: $(DEEPCOPY_TARGETS) k8s-client-gen

image-builder-version: .github/env
	@echo $(GO_VERSION)
	sed -i.bak "s/ARG GOLANG_BUILDER=.*/ARG GOLANG_BUILDER=$(GO_VERSION)/" \
		Dockerfile && rm Dockerfile.bak
	sed -i.bak "s/ARG GOLANG_BUILDER=.*/ARG GOLANG_BUILDER=$(GO_VERSION)/" \
		cmd/prometheus-config-reloader/Dockerfile && rm cmd/prometheus-config-reloader/Dockerfile.bak
	sed -i.bak "s/ARG GOLANG_BUILDER=.*/ARG GOLANG_BUILDER=$(GO_VERSION)/" \
		cmd/admission-webhook/Dockerfile && rm cmd/admission-webhook/Dockerfile.bak

.PHONY: image
image: GOOS := linux # Overriding GOOS value for docker image build
image: operator-image prometheus-config-reloader-image admission-webhook-image

.PHONY: operator-image
operator-image:
	$(CONTAINER_CLI) build --build-arg ARCH=$(ARCH) --build-arg GOARCH=$(GOARCH) --build-arg OS=$(GOOS) -t $(IMAGE_OPERATOR):$(TAG) .

.PHONY: prometheus-config-reloader-image
prometheus-config-reloader-image:
	$(CONTAINER_CLI) build --build-arg ARCH=$(ARCH) --build-arg GOARCH=$(GOARCH) --build-arg OS=$(GOOS) -t $(IMAGE_RELOADER):$(TAG) -f cmd/prometheus-config-reloader/Dockerfile .

.PHONY: admission-webhook-image
admission-webhook-image:
	$(CONTAINER_CLI) build --build-arg ARCH=$(ARCH) --build-arg GOARCH=$(GOARCH) --build-arg OS=$(GOOS) -t $(IMAGE_WEBHOOK):$(TAG) -f cmd/admission-webhook/Dockerfile .

.PHONY: update-go-deps
update-go-deps:
	for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get -u $$m; \
	done
	(cd pkg/client && go get -u ./...)
	(cd pkg/apis/monitoring && go get -u ./...)
	@echo "Don't forget to run 'make tidy'"

##############
# Generating #
##############

.PHONY: tidy
tidy:
	go mod tidy -v
	cd pkg/apis/monitoring && go mod tidy -v -modfile=go.mod
	cd pkg/client && go mod tidy -v -modfile=go.mod
	cd scripts && go mod tidy -v -modfile=go.mod

.PHONY: generate
generate: k8s-gen generate-crds bundle.yaml example/mixin/alerts.yaml example/thanos/thanos.yaml example/admission-webhook example/alertmanager-crd-conversion generate-docs image-builder-version

# For now, the v1beta1 CRDs aren't part of the default bundle because they
# require to deploy/run the conversion webhook.
# They are provided in a separate directory
# (example/prometheus-operator-crd-full) and we generate jsonnet code that can
# be used to patch the "default" jsonnet CRD.
.PHONY: generate-crds
generate-crds: $(CONTROLLER_GEN_BINARY) $(GOJSONTOYAML_BINARY) $(TYPES_V1_TARGET) $(TYPES_V1ALPHA1_TARGET) $(TYPES_V1BETA1_TARGET)
	cd pkg/apis/monitoring && $(CONTROLLER_GEN_BINARY) $(CRD_OPTIONS) paths=./v1/. paths=./v1alpha1/. output:crd:dir=$(PWD)/example/prometheus-operator-crd/
	cd pkg/apis/monitoring && $(CONTROLLER_GEN_BINARY) $(CRD_OPTIONS) paths=./... output:crd:dir=$(PWD)/example/prometheus-operator-crd-full
	VERSION=$(VERSION) ./scripts/generate/append-operator-version.sh
	find example/prometheus-operator-crd/ -name '*.yaml' -print0 | xargs -0 -I{} sh -c '$(GOJSONTOYAML_BINARY) -yamltojson < "$$1" | jq > "$(PWD)/jsonnet/prometheus-operator/$$(basename $$1 | cut -d'_' -f2 | cut -d. -f1)-crd.json"' -- {}
	echo "// Code generated using 'make generate-crds'. DO NOT EDIT." > $(PWD)/jsonnet/prometheus-operator/alertmanagerconfigs-v1beta1-crd.libsonnet
	echo "{spec+: {versions+: $$($(GOJSONTOYAML_BINARY) -yamltojson < example/prometheus-operator-crd-full/monitoring.coreos.com_alertmanagerconfigs.yaml | jq '.spec.versions | map(select(.name == "v1beta1"))')}}" | $(JSONNETFMT_BINARY) - >> $(PWD)/jsonnet/prometheus-operator/alertmanagerconfigs-v1beta1-crd.libsonnet

.PHONY: generate-tls-certs
generate-tls-certs:
	mkdir -p $(CERTS_DIR) && \
	(cd scripts && GOOS=$(OS) GOARCH=$(GOARCH) go run -v ./certs/.)

.PHONY: generate-docs
generate-docs: $(shell find Documentation -type f)

bundle.yaml: generate-crds $(shell find example/rbac/prometheus-operator/*.yaml -type f)
	scripts/generate-bundle.sh

# stripped-down-crds.yaml is a version of the Prometheus Operator CRDs with all
# description fields being removed. It is meant as a workaround for the issue
# that `kubectl apply -f ...` might fail with the full version of the CRDs
# because of too long annotations field.
# See https://github.com/prometheus-operator/prometheus-operator/issues/4355
stripped-down-crds.yaml: $(shell find example/prometheus-operator-crd/*.yaml -type f) $(GOJSONTOYAML_BINARY)
	: > $@
	for f in example/prometheus-operator-crd/*.yaml; do echo '---' >> $@; $(GOJSONTOYAML_BINARY) -yamltojson < $$f | jq 'walk(if type == "object" then with_entries(if .value|type=="object" then . else select(.key | test("description") | not) end) else . end)' | $(GOJSONTOYAML_BINARY) >> $@; done

scripts/generate/vendor: $(JB_BINARY) $(shell find jsonnet/prometheus-operator -type f)
	cd scripts/generate; $(JB_BINARY) install;

example/non-rbac/prometheus-operator.yaml: scripts/generate/vendor VERSION $(shell find jsonnet -type f)
	scripts/generate/build-non-rbac-prometheus-operator.sh

example/mixin/alerts.yaml: $(JSONNET_BINARY) $(GOJSONTOYAML_BINARY)
	-mkdir -p example/alerts
	$(JSONNET_BINARY) jsonnet/mixin/alerts.jsonnet | $(GOJSONTOYAML_BINARY) > $@

RBAC_MANIFESTS = example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml example/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml example/rbac/prometheus-operator/prometheus-operator-service-account.yaml example/rbac/prometheus-operator/prometheus-operator-deployment.yaml
$(RBAC_MANIFESTS): scripts/generate/vendor VERSION $(shell find jsonnet -type f)
	scripts/generate/build-rbac-prometheus-operator.sh

example/thanos/thanos.yaml: scripts/generate/vendor scripts/generate/thanos.jsonnet $(shell find jsonnet -type f)
	scripts/generate/build-thanos-example.sh

example/admission-webhook: scripts/generate/vendor scripts/generate/admission-webhook.jsonnet $(shell find jsonnet -type f)
	scripts/generate/build-admission-webhook-example.sh

example/alertmanager-crd-conversion: scripts/generate/vendor scripts/generate/conversion-webhook-patch-for-alertmanagerconfig-crd.jsonnet $(shell find jsonnet -type f)
	scripts/generate/build-conversion-webhook-patch-for-alertmanagerconfig-crd.sh

FULLY_GENERATED_DOCS = Documentation/api-reference/api.md Documentation/getting-started/compatibility.md Documentation/platform/operator.md

Documentation/platform/operator.md: operator
	$(MDOX_BINARY) fmt $@

Documentation/getting-started/compatibility.md: pkg/operator/defaults.go
	$(MDOX_BINARY) fmt $@

Documentation/api-reference/api.md: $(TYPES_V1_TARGET) $(TYPES_V1ALPHA1_TARGET) $(TYPES_V1BETA1_TARGET)
	GODEBUG=$(GODEBUG) $(API_DOC_GEN_BINARY) -api-dir "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/" -config "$(PWD)/scripts/docs/config.json" -template-dir "$(PWD)/scripts/docs/templates" -out-file "$(PWD)/Documentation/api-reference/api.md"

##############
# Formatting #
##############

.PHONY: format
format: go-fmt jsonnet-fmt check-license shellcheck docs

.PHONY: go-fmt
go-fmt:
	gofmt -s -w .

.PHONY: jsonnet-fmt
jsonnet-fmt: $(JSONNETFMT_BINARY)
	find . -name *.jsonnet -or -name *.libsonnet -not -path "*/vendor/*" -print0 | xargs -0 $(JSONNETFMT_BINARY) -i

.PHONY: check-license
check-license:
	./scripts/check_license.sh

.PHONY: shellcheck
shellcheck: $(SHELLCHECK_BINARY)
	$(SHELLCHECK_BINARY) $(shell find . -type f -name "*.sh" -not -path "*/vendor/*")

.PHONY: check-metrics
check-metrics: $(PROMLINTER_BINARY)
	$(PROMLINTER_BINARY) lint .

.PHONY: check
check: check-golang check-api

.PHONY: check-golang
check-golang: $(GOLANGCILINTER_BINARY)
	$(GOLANGCILINTER_BINARY) run -v

.PHONY: check-api
check-api: $(GOLANGCIKUBEAPILINTER_BINARY)
	cd pkg/apis/monitoring && $(GOLANGCIKUBEAPILINTER_BINARY) run -v --config $(ROOT_DIR)/.golangci-kal.yml

.PHONY: check
fix: fix-golang fix-api

.PHONY: fix-golang
fix-golang: $(GOLANGCILINTER_BINARY)
	$(GOLANGCILINTER_BINARY) run --fix

.PHONY: fix-api
fix-api: $(GOLANGCIKUBEAPILINTER_BINARY)
	cd pkg/apis/monitoring && $(GOLANGCIKUBEAPILINTER_BINARY) run -v --config $(ROOT_DIR)/.golangci-kal.yml --fix

MDOX_VALIDATE_CONFIG?=.mdox.validate.yaml
MD_FILES_TO_FORMAT=$(filter-out $(FULLY_GENERATED_DOCS), $(shell find Documentation -name "*.md")) $(filter-out ADOPTERS.md, $(shell ls *.md))

.PHONY: docs
docs: $(MDOX_BINARY)
	@echo ">> formatting and local/remote link check"
	GITHUB_TOKEN=$(GITHUB_TOKEN) $(MDOX_BINARY) fmt --soft-wraps -l --links.localize.address-regex="https://prometheus-operator.dev/.*" --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)

.PHONY: check-docs
check-docs: $(MDOX_BINARY)
	@echo ">> checking formatting and local/remote links"
	GITHUB_TOKEN=$(GITHUB_TOKEN) $(MDOX_BINARY) fmt --soft-wraps --check -l --links.localize.address-regex="https://prometheus-operator.dev/.*" --links.validate.config-file=$(MDOX_VALIDATE_CONFIG) $(MD_FILES_TO_FORMAT)

###########
# Testing #
###########

.PHONY: test
test: test-unit test-long test-e2e

.PHONY: test-unit
test-unit: test-prometheus-goldenfiles
	go test -race $(TEST_RUN_ARGS) -short $(pkgs) -count=1 -v

.PHONY: test-long
test-long: test-prometheus-goldenfiles
	go test $(TEST_RUN_ARGS) $(pkgs) -count=1 -v

.PHONY: test-unit-update-golden
test-unit-update-golden:
	./scripts/update-golden-files.sh

.PHONY: test-prometheus-goldenfiles
test-prometheus-goldenfiles: $(PROMTOOL_BINARY)
	$(PROMTOOL_BINARY) check config --syntax-only pkg/prometheus/testdata/*.golden

test/instrumented-sample-app/certs/cert.pem test/instrumented-sample-app/certs/key.pem:
	cd test/instrumented-sample-app && make generate-certs

$(CERTS_DIR)/ca.key $(CERTS_DIR)/ca.crt $(CERTS_DIR)/client.key $(CERTS_DIR)/client.crt $(CERTS_DIR)/bad_ca.key $(CERTS_DIR)/bad_ca.crt $(CERTS_DIR)/bad_client.key $(CERTS_DIR)/bad_client.crt:
	$(MAKE) generate-tls-certs

.PHONY: test-e2e
test-e2e: KUBECONFIG?=$(HOME)/.kube/config
test-e2e: test/instrumented-sample-app/certs/cert.pem test/instrumented-sample-app/certs/key.pem
	go test -timeout 120m -v ./test/e2e/ $(TEST_RUN_ARGS) --kubeconfig=$(KUBECONFIG) --operator-image=$(IMAGE_OPERATOR):$(TAG) -count=1

.PHONY: test-e2e-alertmanager
test-e2e-alertmanager:
	EXCLUDE_ALERTMANAGER_TESTS= EXCLUDE_PROMETHEUS_TESTS=exclude EXCLUDE_PROMETHEUS_ALL_NS_TESTS=exclude EXCLUDE_THANOSRULER_TESTS=exclude EXCLUDE_OPERATOR_UPGRADE_TESTS=exclude EXCLUDE_FEATURE_GATED_TESTS=exclude EXCLUDE_PROMETHEUS_UPGRADE_TESTS=exclude $(MAKE) test-e2e

.PHONY: test-e2e-prometheus
test-e2e-prometheus:
	EXCLUDE_ALERTMANAGER_TESTS=exclude EXCLUDE_PROMETHEUS_TESTS= EXCLUDE_PROMETHEUS_ALL_NS_TESTS=exclude EXCLUDE_THANOSRULER_TESTS=exclude EXCLUDE_OPERATOR_UPGRADE_TESTS=exclude EXCLUDE_FEATURE_GATED_TESTS=exclude EXCLUDE_PROMETHEUS_UPGRADE_TESTS=exclude $(MAKE) test-e2e

.PHONY: test-e2e-prometheus-all-namespaces
test-e2e-prometheus-all-namespaces:
	EXCLUDE_ALERTMANAGER_TESTS=exclude EXCLUDE_PROMETHEUS_TESTS=exclude EXCLUDE_PROMETHEUS_ALL_NS_TESTS= EXCLUDE_THANOSRULER_TESTS=exclude EXCLUDE_OPERATOR_UPGRADE_TESTS=exclude EXCLUDE_FEATURE_GATED_TESTS=exclude EXCLUDE_PROMETHEUS_UPGRADE_TESTS=exclude $(MAKE) test-e2e

.PHONY: test-e2e-thanos-ruler
test-e2e-thanos-ruler:
	EXCLUDE_ALERTMANAGER_TESTS=exclude EXCLUDE_PROMETHEUS_TESTS=exclude EXCLUDE_PROMETHEUS_ALL_NS_TESTS=exclude EXCLUDE_THANOSRULER_TESTS= EXCLUDE_OPERATOR_UPGRADE_TESTS=exclude EXCLUDE_FEATURE_GATED_TESTS=exclude EXCLUDE_PROMETHEUS_UPGRADE_TESTS=exclude $(MAKE) test-e2e

.PHONY: test-e2e-operator-upgrade
test-e2e-operator-upgrade:
	EXCLUDE_ALERTMANAGER_TESTS=exclude EXCLUDE_PROMETHEUS_TESTS=exclude EXCLUDE_PROMETHEUS_ALL_NS_TESTS=exclude EXCLUDE_THANOSRULER_TESTS=exclude EXCLUDE_OPERATOR_UPGRADE_TESTS= EXCLUDE_FEATURE_GATED_TESTS=exclude EXCLUDE_PROMETHEUS_UPGRADE_TESTS=exclude $(MAKE) test-e2e

.PHONY: test-e2e-prometheus-upgrade
test-e2e-prometheus-upgrade:
	EXCLUDE_ALERTMANAGER_TESTS=exclude EXCLUDE_PROMETHEUS_TESTS=exclude EXCLUDE_PROMETHEUS_ALL_NS_TESTS=exclude EXCLUDE_THANOSRULER_TESTS=exclude EXCLUDE_OPERATOR_UPGRADE_TESTS=exclude EXCLUDE_FEATURE_GATED_TESTS=exclude EXCLUDE_PROMETHEUS_UPGRADE_TESTS= $(MAKE) test-e2e

.PHONY: test-e2e-feature-gates
test-e2e-feature-gates:
	EXCLUDE_ALERTMANAGER_TESTS=exclude EXCLUDE_PROMETHEUS_TESTS=exclude EXCLUDE_PROMETHEUS_ALL_NS_TESTS=exclude EXCLUDE_THANOSRULER_TESTS=exclude EXCLUDE_OPERATOR_UPGRADE_TESTS=exclude EXCLUDE_FEATURE_GATED_TESTS= EXCLUDE_PROMETHEUS_UPGRADE_TESTS=exclude $(MAKE) test-e2e

.PHONY: test-e2e-images
test-e2e-images: image $(TOOLS_BIN_DIR)
ifeq (podman, $(CONTAINER_CLI))
	podman save --quiet -o $(TOOLS_BIN_DIR)/prometheus-operator.tar $(IMAGE_OPERATOR):$(TAG)
	podman save --quiet -o $(TOOLS_BIN_DIR)/prometheus-config-reloader.tar $(IMAGE_RELOADER):$(TAG)
	podman save --quiet -o $(TOOLS_BIN_DIR)/admission-webhook.tar $(IMAGE_WEBHOOK):$(TAG)
	kind load image-archive -n $(KIND_CONTEXT) $(TOOLS_BIN_DIR)/prometheus-operator.tar
	kind load image-archive -n $(KIND_CONTEXT) $(TOOLS_BIN_DIR)/prometheus-config-reloader.tar
	kind load image-archive -n $(KIND_CONTEXT) $(TOOLS_BIN_DIR)/admission-webhook.tar
else
	kind load docker-image -n $(KIND_CONTEXT) $(IMAGE_OPERATOR):$(TAG)
	kind load docker-image -n $(KIND_CONTEXT) $(IMAGE_RELOADER):$(TAG)
	kind load docker-image -n $(KIND_CONTEXT) $(IMAGE_WEBHOOK):$(TAG)
endif

############
# Binaries #
############

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

$(TOOLING): $(TOOLS_BIN_DIR)
	@echo Installing tools from scripts/tools.go
	@cat scripts/tools.go | grep _ | awk -F'"' '{print $$2}' | GOBIN=$(TOOLS_BIN_DIR) xargs -tI % go install -mod=readonly -modfile=scripts/go.mod %
	@GOBIN=$(TOOLS_BIN_DIR) go install $(GO_PKG)/cmd/po-docgen
	@GOBIN=$(TOOLS_BIN_DIR) $(GOLANGCILINTER_BINARY) custom
	@echo Downloading shellcheck
	@cd $(TOOLS_BIN_DIR) && wget -qO- "https://github.com/koalaman/shellcheck/releases/download/stable/shellcheck-stable.$(GOOS).x86_64.tar.xz" | tar -xJv --strip=1 shellcheck-stable/shellcheck

# generate k8s generator variable and target,
# i.e. if $(1)=informer-gen:
#
# INFORMER_GEN_BINARY=/home/user/go/bin/informer-gen
#
# /home/user/go/bin/informer-gen:
#	go get -u -d k8s.io/code-generator/cmd/informer-gen
#	cd /home/user/go/src/k8s.io/code-generator; git checkout release-1.14
#	go install k8s.io/code-generator/cmd/informer-gen
#
define _K8S_GEN_VAR_TARGET_
$(shell echo $(1) | tr '[:lower:]' '[:upper:]' | tr '-' '_')_BINARY:=$(TOOLS_BIN_DIR)/$(1)

$(TOOLS_BIN_DIR)/$(1):
	@GOBIN=$(TOOLS_BIN_DIR) go install -mod=readonly -modfile=scripts/go.mod k8s.io/code-generator/cmd/$(1)

endef

$(foreach binary,$(K8S_GEN_BINARIES),$(eval $(call _K8S_GEN_VAR_TARGET_,$(binary))))

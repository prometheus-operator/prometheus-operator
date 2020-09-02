SHELL=/bin/bash -o pipefail

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
ifeq ($(GOARCH),arm)
	ARCH=armv7
else
	ARCH=$(GOARCH)
endif

GO_PKG=github.com/prometheus-operator/prometheus-operator
REPO?=quay.io/prometheus-operator/prometheus-operator
REPO_PROMETHEUS_CONFIG_RELOADER?=quay.io/prometheus-operator/prometheus-config-reloader
REPO_PROMETHEUS_OPERATOR_LINT?=quay.io/prometheus-operator/prometheus-operator-lint
TAG?=$(shell git rev-parse --short HEAD)
VERSION?=$(shell cat VERSION | tr -d " \t\n\r")

TYPES_V1_TARGET := pkg/apis/monitoring/v1/types.go
TYPES_V1_TARGET += pkg/apis/monitoring/v1/thanos_types.go

TOOLS_BIN_DIR ?= $(shell pwd)/tmp/bin
export PATH := $(TOOLS_BIN_DIR):$(PATH)

PO_DOCGEN_BINARY:=$(TOOLS_BIN_DIR)/po-docgen
CONTROLLER_GEN_BINARY := $(TOOLS_BIN_DIR)/controller-gen
EMBEDMD_BINARY=$(TOOLS_BIN_DIR)/embedmd
JB_BINARY=$(TOOLS_BIN_DIR)/jb
GOJSONTOYAML_BINARY=$(TOOLS_BIN_DIR)/gojsontoyaml
JSONNET_BINARY=$(TOOLS_BIN_DIR)/jsonnet
JSONNETFMT_BINARY=$(TOOLS_BIN_DIR)/jsonnetfmt
SHELLCHECK_BINARY=$(TOOLS_BIN_DIR)/shellcheck
TOOLING=$(PO_DOCGEN_BINARY) $(CONTROLLER_GEN_BINARY) $(EMBEDMD_BINARY) $(GOBINDATA_BINARY) $(JB_BINARY) $(GOJSONTOYAML_BINARY) $(JSONNET_BINARY) $(JSONNETFMT_BINARY) $(SHELLCHECK_BINARY)

K8S_GEN_VERSION:=release-1.14
K8S_GEN_BINARIES:=informer-gen lister-gen client-gen
K8S_GEN_ARGS:=--go-header-file $(shell pwd)/.header --v=1 --logtostderr

K8S_GEN_DEPS:=.header
K8S_GEN_DEPS+=$(TYPES_V1_TARGET)
K8S_GEN_DEPS+=$(foreach bin,$(K8S_GEN_BINARIES),$(TOOLS_BIN_DIR)/$(bin))

GO_BUILD_RECIPE=GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags="-s -X $(GO_PKG)/pkg/version.Version=$(VERSION)"
pkgs = $(shell go list ./... | grep -v /test/ | grep -v /contrib/)

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
build: operator prometheus-config-reloader k8s-gen po-lint

.PHONY: operator
operator:
	$(GO_BUILD_RECIPE) -o $@ cmd/operator/main.go

.PHONY: prometheus-config-reloader
prometheus-config-reloader:
	$(GO_BUILD_RECIPE) -o $@ cmd/$@/main.go

.PHONY: po-lint
po-lint:
	$(GO_BUILD_RECIPE) -o po-lint cmd/po-lint/main.go

DEEPCOPY_TARGET := pkg/apis/monitoring/v1/zz_generated.deepcopy.go
$(DEEPCOPY_TARGET): $(CONTROLLER_GEN_BINARY)
	cd ./pkg/apis/monitoring/v1 && $(CONTROLLER_GEN_BINARY) object:headerFile=$(CURDIR)/.header \
		paths=.

CLIENT_TARGET := pkg/client/versioned/clientset.go
$(CLIENT_TARGET): $(K8S_GEN_DEPS)
	$(CLIENT_GEN_BINARY) \
	$(K8S_GEN_ARGS) \
	--input-base     "" \
	--clientset-name "versioned" \
	--input	         "$(GO_PKG)/pkg/apis/monitoring/v1" \
	--output-package "$(GO_PKG)/pkg/client"

LISTER_TARGET := pkg/client/listers/monitoring/v1/prometheus.go
$(LISTER_TARGET): $(K8S_GEN_DEPS)
	$(LISTER_GEN_BINARY) \
	$(K8S_GEN_ARGS) \
	--input-dirs     "$(GO_PKG)/pkg/apis/monitoring/v1" \
	--output-package "$(GO_PKG)/pkg/client/listers"

INFORMER_TARGET := pkg/client/informers/externalversions/monitoring/v1/prometheus.go
$(INFORMER_TARGET): $(K8S_GEN_DEPS) $(LISTER_TARGET) $(CLIENT_TARGET)
	$(INFORMER_GEN_BINARY) \
	$(K8S_GEN_ARGS) \
	--versioned-clientset-package "$(GO_PKG)/pkg/client/versioned" \
	--listers-package "$(GO_PKG)/pkg/client/listers" \
	--input-dirs      "$(GO_PKG)/pkg/apis/monitoring/v1" \
	--output-package  "$(GO_PKG)/pkg/client/informers"

.PHONY: k8s-gen
k8s-gen: \
	$(DEEPCOPY_TARGET) \
	$(CLIENT_TARGET) \
	$(LISTER_TARGET) \
	$(INFORMER_TARGET) \
	$(OPENAPI_TARGET)

.PHONY: image
image: .hack-operator-image .hack-prometheus-config-reloader-image

.hack-operator-image: Dockerfile operator
# Create empty target file, for the sole purpose of recording when this target
# was last executed via the last-modification timestamp on the file. See
# https://www.gnu.org/software/make/manual/make.html#Empty-Targets
	docker build --build-arg ARCH=$(ARCH) --build-arg OS=$(GOOS) -t $(REPO):$(TAG) .
	touch $@

.hack-prometheus-config-reloader-image: cmd/prometheus-config-reloader/Dockerfile prometheus-config-reloader
# Create empty target file, for the sole purpose of recording when this target
# was last executed via the last-modification timestamp on the file. See
# https://www.gnu.org/software/make/manual/make.html#Empty-Targets
	docker build --build-arg ARCH=$(ARCH) --build-arg OS=$(GOOS) -t $(REPO_PROMETHEUS_CONFIG_RELOADER):$(TAG) -f cmd/prometheus-config-reloader/Dockerfile .
	touch $@


##############
# Generating #
##############

.PHONY: tidy
tidy:
	go mod tidy -v
	cd scripts && go mod tidy -v -modfile=go.mod

.PHONY: generate
generate: $(DEEPCOPY_TARGET) generate-crds bundle.yaml example/mixin/alerts.yaml $(shell find Documentation -type f)

.PHONY: generate-crds
generate-crds: $(CONTROLLER_GEN_BINARY) $(GOJSONTOYAML_BINARY) $(TYPES_V1_TARGET)
	GOOS=$(OS) GOARCH=$(ARCH) go run -v ./scripts/generate-crds.go --controller-gen=$(CONTROLLER_GEN_BINARY) --gojsontoyaml=$(GOJSONTOYAML_BINARY)

.PHONY: generate-remote-write-certs
generate-remote-write-certs:
	mkdir -p test/e2e/remote_write_certs && \
	openssl req -newkey rsa:4096 \
		-new -nodes -x509 -sha256 \
		-days 3650 \
		-out test/e2e/remote_write_certs/ca.crt \
		-keyout test/e2e/remote_write_certs/ca.key \
		-subj "/CN=caandserver.com"
	openssl req -new -newkey rsa:4096 \
		-keyout test/e2e/remote_write_certs/client.key \
		-out test/e2e/remote_write_certs/client.csr \
		-nodes \
		-subj "/CN=PrometheusRemoteWriteClient"
	openssl x509 -req -sha256 \
		-days 3650 \
		-in test/e2e/remote_write_certs/client.csr \
		-CA test/e2e/remote_write_certs/ca.crt \
		-CAkey test/e2e/remote_write_certs/ca.key \
		-set_serial 02 \
		-out test/e2e/remote_write_certs/client.crt
	rm test/e2e/remote_write_certs/client.csr
	openssl req -newkey rsa:4096 \
		-new -nodes -x509 -sha256 \
		-days 3650 \
		-out test/e2e/remote_write_certs/bad_ca.crt \
		-keyout test/e2e/remote_write_certs/bad_ca.key \
		-subj "/CN=badcaandserver.com"
	openssl req -new -newkey rsa:4096 \
		-keyout test/e2e/remote_write_certs/bad_client.key \
		-out test/e2e/remote_write_certs/bad_client.csr \
		-nodes \
		-subj "/CN=BadPrometheusRemoteWriteClient"
	openssl x509 -req -sha256 \
		-days 3650 \
		-in test/e2e/remote_write_certs/bad_client.csr \
		-CA test/e2e/remote_write_certs/bad_ca.crt \
		-CAkey test/e2e/remote_write_certs/bad_ca.key \
		-set_serial 02 \
		-out test/e2e/remote_write_certs/bad_client.crt
	rm test/e2e/remote_write_certs/bad_client.csr

bundle.yaml: generate-crds $(shell find example/rbac/prometheus-operator/*.yaml -type f)
	scripts/generate-bundle.sh

scripts/generate/vendor: $(JB_BINARY) $(shell find jsonnet/prometheus-operator -type f)
	cd scripts/generate; $(JB_BINARY) install;

example/non-rbac/prometheus-operator.yaml: scripts/generate/vendor scripts/generate/prometheus-operator-non-rbac.jsonnet $(shell find jsonnet -type f)
	scripts/generate/build-non-rbac-prometheus-operator.sh

example/mixin/alerts.yaml: $(JSONNET_BINARY) $(GOJSONTOYAML_BINARY)
	-mkdir -p example/alerts
	$(JSONNET_BINARY) jsonnet/mixin/alerts.jsonnet | $(GOJSONTOYAML_BINARY) > $@

RBAC_MANIFESTS = example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml example/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml example/rbac/prometheus-operator/prometheus-operator-service-account.yaml example/rbac/prometheus-operator/prometheus-operator-deployment.yaml
$(RBAC_MANIFESTS): scripts/generate/vendor scripts/generate/prometheus-operator-rbac.jsonnet $(shell find jsonnet -type f)
	scripts/generate/build-rbac-prometheus-operator.sh

jsonnet/prometheus-operator/prometheus-operator.libsonnet: VERSION
	# note: use temporary file to preserve compatibility with darwin
	sed -i.bak \
		"s/prometheusOperator: 'v.*',/prometheusOperator: 'v$(VERSION)',/" \
		jsonnet/prometheus-operator/prometheus-operator.libsonnet;
	rm jsonnet/prometheus-operator/prometheus-operator.libsonnet.bak

FULLY_GENERATED_DOCS = Documentation/api.md Documentation/compatibility.md
TO_BE_EXTENDED_DOCS = $(filter-out $(FULLY_GENERATED_DOCS), $(shell find Documentation -type f))

Documentation/api.md: $(PO_DOCGEN_BINARY) $(TYPES_V1_TARGET)
	$(PO_DOCGEN_BINARY) api $(TYPES_V1_TARGET) > $@

Documentation/compatibility.md: $(PO_DOCGEN_BINARY) pkg/prometheus/statefulset.go
	$(PO_DOCGEN_BINARY) compatibility > $@

$(TO_BE_EXTENDED_DOCS): $(EMBEDMD_BINARY) $(shell find example) bundle.yaml
	$(EMBEDMD_BINARY) -w `find Documentation -name "*.md"`


##############
# Formatting #
##############

.PHONY: format
format: go-fmt jsonnet-fmt check-license shellcheck

.PHONY: go-fmt
go-fmt:
	go fmt $(pkgs)

.PHONY: jsonnet-fmt
jsonnet-fmt: $(JSONNETFMT_BINARY)
	# *.*sonnet will match *.jsonnet and *.libsonnet files but nothing else in this repository
	find . -name *.jsonnet -not -path "*/vendor/*" -print0 | xargs -0 $(JSONNETFMT_BINARY) -i

.PHONY: check-license
check-license:
	./scripts/check_license.sh

.PHONY: shellcheck
shellcheck: $(SHELLCHECK_BINARY)
	$(SHELLCHECK_BINARY) $(shell find . -type f -name "*.sh" -not -path "*/vendor/*")

###########
# Testing #
###########

.PHONY: test
test: test-unit test-long test-e2e

.PHONY: test-unit
test-unit:
	go test -race $(TEST_RUN_ARGS) -short $(pkgs) -count=1 -v

.PHONY: test-long
test-long:
	go test $(TEST_RUN_ARGS) $(pkgs) -count=1 -v

test/instrumented-sample-app/certs/cert.pem test/instrumented-sample-app/certs/key.pem:
	cd test/instrumented-sample-app && make generate-certs

test/e2e/remote_write_certs/ca.key test/e2e/remote_write_certs/ca.crt test/e2e/remote_write_certs/client.key test/e2e/remote_write_certs/client.crt test/e2e/remote_write_certs/bad_ca.key test/e2e/remote_write_certs/bad_ca.crt test/e2e/remote_write_certs/bad_client.key test/e2e/remote_write_certs/bad_client.crt:
	make generate-remote-write-certs

.PHONY: test-e2e
test-e2e: KUBECONFIG?=$(HOME)/.kube/config
test-e2e: test/instrumented-sample-app/certs/cert.pem test/instrumented-sample-app/certs/key.pem
	go test -timeout 55m -v ./test/e2e/ $(TEST_RUN_ARGS) --kubeconfig=$(KUBECONFIG) --operator-image=$(REPO):$(TAG) -count=1

############
# Binaries #
############

$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

$(TOOLING): $(TOOLS_BIN_DIR)
	@echo Installing tools from scripts/tools.go
	@cat scripts/tools.go | grep _ | awk -F'"' '{print $$2}' | GOBIN=$(TOOLS_BIN_DIR) xargs -tI % go install -mod=readonly -modfile=scripts/go.mod %
	@GOBIN=$(TOOLS_BIN_DIR) go install $(GO_PKG)/cmd/po-docgen
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
	@go install -mod=readonly -modfile=scripts/go.mod k8s.io/code-generator/cmd/$(1)

endef

$(foreach binary,$(K8S_GEN_BINARIES),$(eval $(call _K8S_GEN_VAR_TARGET_,$(binary))))

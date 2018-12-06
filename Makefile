SHELL=/bin/bash -o pipefail

GO_PKG=github.com/coreos/prometheus-operator
REPO?=quay.io/coreos/prometheus-operator
REPO_PROMETHEUS_CONFIG_RELOADER?=quay.io/coreos/prometheus-config-reloader
TAG?=$(shell git rev-parse --short HEAD)

FIRST_GOPATH:=$(firstword $(subst :, ,$(shell go env GOPATH)))
PO_CRDGEN_BINARY:=$(FIRST_GOPATH)/bin/po-crdgen
OPENAPI_GEN_BINARY:=$(FIRST_GOPATH)/bin/openapi-gen
GOJSONTOYAML_BINARY:=$(FIRST_GOPATH)/bin/gojsontoyaml
JB_BINARY:=$(FIRST_GOPATH)/bin/jb
PO_DOCGEN_BINARY:=$(FIRST_GOPATH)/bin/po-docgen
EMBEDMD_BINARY:=$(FIRST_GOPATH)/bin/embedmd

TYPES_V1_TARGET:=pkg/apis/monitoring/v1/types.go

# Unfortunately kube-openapi doesn't seem to be properly tagged yet as the other generator binary.
# Starting with https://github.com/kubernetes/kube-openapi/commit/07437455b254b00a4deb3b420e790b2215450487
# type object declarations are added which break prometheus operator.
#
# TODO(sur): bump this to a proper release branch once upstream resolved this.
K8S_OPENAPI_GEN_VERSION:=0317810137be915b9cf888946c6e115c1bfac693

K8S_GEN_VERSION:=release-1.13
K8S_GEN_BINARIES:=deepcopy-gen informer-gen lister-gen client-gen
K8S_GEN_ARGS:=--go-header-file $(FIRST_GOPATH)/src/$(GO_PKG)/.header --v=1 --logtostderr

K8S_GEN_DEPS:=.header
K8S_GEN_DEPS+=$(TYPES_V1_TARGET)
K8S_GEN_DEPS+=$(foreach bin,$(K8S_GEN_BINARIES),$(FIRST_GOPATH)/bin/$(bin))
K8S_GEN_DEPS+=$(OPENAPI_GEN_BINARY)

GOLANG_FILES:=$(shell find . -name \*.go -print)
pkgs = $(shell go list ./... | grep -v /vendor/ | grep -v /test/ | grep -v /contrib/)

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
build: operator prometheus-config-reloader k8s-gen

.PHONY: operator
operator: $(GOLANG_FILES)
	GOOS=linux CGO_ENABLED=0 go build \
	-ldflags "-X $(GO_PKG)/pkg/version.Version=$(shell cat VERSION)" \
	-o $@ cmd/operator/main.go

.PHONY: prometheus-config-reloader
prometheus-config-reloader:
	GOOS=linux CGO_ENABLED=0 go build \
	-ldflags "-X $(GO_PKG)/pkg/version.Version=$(shell cat VERSION)" \
	-o $@ cmd/$@/main.go

DEEPCOPY_TARGET := pkg/apis/monitoring/v1/zz_generated.deepcopy.go
$(DEEPCOPY_TARGET): $(K8S_GEN_DEPS)
	$(DEEPCOPY_GEN_BINARY) \
	$(K8S_GEN_ARGS) \
	--input-dirs    "$(GO_PKG)/pkg/apis/monitoring/v1" \
	--bounding-dirs "$(GO_PKG)/pkg/apis/monitoring" \
	--output-file-base zz_generated.deepcopy

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

OPENAPI_TARGET := pkg/apis/monitoring/v1/openapi_generated.go
$(OPENAPI_TARGET): $(K8S_GEN_DEPS)
	$(OPENAPI_GEN_BINARY) \
	$(K8S_GEN_ARGS) \
	-i $(GO_PKG)/pkg/apis/monitoring/v1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/api/core/v1 \
	-p $(GO_PKG)/pkg/apis/monitoring/v1

.PHONY: k8s-gen
k8s-gen: \
 $(DEEPCOPY_TARGET) \
 $(CLIENT_TARGET) \
 $(LISTER_TARGET) \
 $(INFORMER_TARGET) \
 $(OPENAPI_TARGET)

.PHONY: image
image: hack/operator-image hack/prometheus-config-reloader-image

hack/operator-image: Dockerfile operator
# Create empty target file, for the sole purpose of recording when this target
# was last executed via the last-modification timestamp on the file. See
# https://www.gnu.org/software/make/manual/make.html#Empty-Targets
	docker build -t $(REPO):$(TAG) .
	touch $@

hack/prometheus-config-reloader-image: cmd/prometheus-config-reloader/Dockerfile prometheus-config-reloader
# Create empty target file, for the sole purpose of recording when this target
# was last executed via the last-modification timestamp on the file. See
# https://www.gnu.org/software/make/manual/make.html#Empty-Targets
	docker build -t $(REPO_PROMETHEUS_CONFIG_RELOADER):$(TAG) -f cmd/prometheus-config-reloader/Dockerfile .
	touch $@


##############
# Generating #
##############

.PHONY: generate
generate: $(DEEPCOPY_TARGET) $(OPENAPI_TARGET) $(shell find jsonnet/prometheus-operator/*-crd.libsonnet -type f) bundle.yaml kube-prometheus $(shell find Documentation -type f)

.PHONY: generate-in-docker
generate-in-docker: hack/jsonnet-docker-image
	hack/generate-in-docker.sh $(MFLAGS) # MFLAGS are the parent make call's flags

.PHONY: kube-prometheus
kube-prometheus:
	cd contrib/kube-prometheus && $(MAKE) $(MFLAGS) generate

example/prometheus-operator-crd/**.crd.yaml: $(OPENAPI_TARGET) $(PO_CRDGEN_BINARY)
	po-crdgen prometheus > example/prometheus-operator-crd/prometheus.crd.yaml
	po-crdgen alertmanager > example/prometheus-operator-crd/alertmanager.crd.yaml
	po-crdgen servicemonitor > example/prometheus-operator-crd/servicemonitor.crd.yaml
	po-crdgen prometheusrule > example/prometheus-operator-crd/prometheusrule.crd.yaml

jsonnet/prometheus-operator/**-crd.libsonnet: $(shell find example/prometheus-operator-crd/*.crd.yaml -type f) $(GOJSONTOYAML_BINARY)
	cat example/prometheus-operator-crd/alertmanager.crd.yaml   | gojsontoyaml -yamltojson > jsonnet/prometheus-operator/alertmanager-crd.libsonnet
	cat example/prometheus-operator-crd/prometheus.crd.yaml     | gojsontoyaml -yamltojson > jsonnet/prometheus-operator/prometheus-crd.libsonnet
	cat example/prometheus-operator-crd/servicemonitor.crd.yaml | gojsontoyaml -yamltojson > jsonnet/prometheus-operator/servicemonitor-crd.libsonnet
	cat example/prometheus-operator-crd/prometheusrule.crd.yaml | gojsontoyaml -yamltojson > jsonnet/prometheus-operator/prometheusrule-crd.libsonnet

bundle.yaml: $(shell find example/rbac/prometheus-operator/*.yaml -type f)
	hack/generate-bundle.sh

hack/generate/vendor: $(JB_BINARY) $(shell find jsonnet/prometheus-operator -type f)
	cd hack/generate; $(JB_BINARY) install;

example/non-rbac/prometheus-operator.yaml: hack/generate/vendor hack/generate/prometheus-operator-non-rbac.jsonnet $(shell find jsonnet -type f)
	hack/generate/build-non-rbac-prometheus-operator.sh

RBAC_MANIFESTS = example/rbac/prometheus-operator/prometheus-operator-cluster-role.yaml example/rbac/prometheus-operator/prometheus-operator-cluster-role-binding.yaml example/rbac/prometheus-operator/prometheus-operator-service-account.yaml example/rbac/prometheus-operator/prometheus-operator-deployment.yaml
$(RBAC_MANIFESTS): hack/generate/vendor hack/generate/prometheus-operator-rbac.jsonnet $(shell find jsonnet -type f)
	hack/generate/build-rbac-prometheus-operator.sh

jsonnet/prometheus-operator/prometheus-operator.libsonnet: VERSION
	sed -i                                                            \
		"s/prometheusOperator: 'v.*',/prometheusOperator: 'v$(shell cat VERSION)',/" \
		jsonnet/prometheus-operator/prometheus-operator.libsonnet;

FULLY_GENERATED_DOCS = Documentation/api.md Documentation/compatibility.md
TO_BE_EXTENDED_DOCS = $(filter-out $(FULLY_GENERATED_DOCS), $(wildcard Documentation/*.md))

Documentation/api.md: $(PO_DOCGEN_BINARY) $(TYPES_V1_TARGET)
	$(PO_DOCGEN_BINARY) api $(TYPES_V1_TARGET) > $@

Documentation/compatibility.md: $(PO_DOCGEN_BINARY) pkg/prometheus/statefulset.go
	$(PO_DOCGEN_BINARY) compatibility > $@

$(TO_BE_EXTENDED_DOCS): $(EMBEDMD_BINARY) $(shell find example) kube-prometheus
	$(EMBEDMD_BINARY) -w `find Documentation -name "*.md" | grep -v vendor`


##############
# Formatting #
##############

.PHONY: format
format: go-fmt check-license shellcheck

.PHONY: go-fmt
go-fmt:
	go fmt $(pkgs)

.PHONY: check-license
check-license:
	./scripts/check_license.sh

.PHONY: shellcheck
shellcheck:
	docker run -v "${PWD}:/mnt" koalaman/shellcheck:stable $(shell find . -type f -name "*.sh" -not -path "*vendor*")


###########
# Testing #
###########

.PHONY: test
test: test-unit test-e2e

.PHONY: test-unit
test-unit:
	@go test -race $(TEST_RUN_ARGS) -short $(pkgs)

.PHONY: test-e2e
test-e2e: KUBECONFIG?=$(HOME)/.kube/config
test-e2e:
	go test -timeout 55m -v ./test/e2e/ $(TEST_RUN_ARGS) --kubeconfig=$(KUBECONFIG) --operator-image=$(REPO):$(TAG)

.PHONY: test-e2e-helm
test-e2e-helm:
	./helm/hack/e2e-test.sh
	# package the chart and verify if they have the version bumped
	helm/hack/helm-package.sh "alertmanager grafana prometheus prometheus-operator exporter-kube-dns exporter-kube-scheduler exporter-kubelets exporter-node exporter-kube-controller-manager exporter-kube-etcd exporter-kube-state exporter-kubernetes exporter-coredns"
	helm/hack/sync-repo.sh false


########
# Misc #
########

hack/jsonnet-docker-image: scripts/jsonnet/Dockerfile
	docker build -f scripts/jsonnet/Dockerfile -t po-jsonnet .
	touch $@

.PHONY: helm-sync-s3
helm-sync-s3:
	helm/hack/helm-package.sh "alertmanager grafana prometheus prometheus-operator exporter-kube-dns exporter-kube-scheduler exporter-kubelets exporter-node exporter-kube-controller-manager exporter-kube-etcd exporter-kube-state exporter-kubernetes exporter-coredns"
	helm/hack/sync-repo.sh true
	helm/hack/helm-package.sh kube-prometheus
	helm/hack/sync-repo.sh true

############
# Binaries #
############

# generate k8s generator variable and target,
# i.e. if $(1)=informer-gen:
#
# INFORMER_GEN_BINARY=/home/user/go/bin/informer-gen
#
# /home/user/go/bin/informer-gen:
#	go get -u -d k8s.io/code-generator/cmd/informer-gen
#	cd /home/user/go/src/k8s.io/code-generator; git checkout release-1.13
#	go install k8s.io/code-generator/cmd/informer-gen
#
define _K8S_GEN_VAR_TARGET_
$(shell echo $(1) | tr '[:lower:]' '[:upper:]' | tr '-' '_')_BINARY:=$(FIRST_GOPATH)/bin/$(1)

$(FIRST_GOPATH)/bin/$(1):
	go get -u -d k8s.io/code-generator/cmd/$(1)
	cd $(FIRST_GOPATH)/src/k8s.io/code-generator; git checkout $(K8S_GEN_VERSION)
	go install k8s.io/code-generator/cmd/$(1)

endef

$(OPENAPI_GEN_BINARY):
	go get -u -d k8s.io/kube-openapi/cmd/openapi-gen
	cd $(FIRST_GOPATH)/src/k8s.io/kube-openapi; git checkout $(K8S_OPENAPI_GEN_VERSION)
	go install k8s.io/kube-openapi/cmd/openapi-gen

$(foreach binary,$(K8S_GEN_BINARIES),$(eval $(call _K8S_GEN_VAR_TARGET_,$(binary))))

$(EMBEDMD_BINARY):
	@go get github.com/campoy/embedmd

$(JB_BINARY):
	go get -u github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb

$(PO_CRDGEN_BINARY): cmd/po-crdgen/main.go $(OPENAPI_TARGET)
	go install $(GO_PKG)/cmd/po-crdgen

$(PO_DOCGEN_BINARY): $(shell find cmd/po-docgen -type f) $(TYPES_V1_TARGET)
	go install $(GO_PKG)/cmd/po-docgen

$(GOJSONTOYAML_BINARY):
	go get -u github.com/brancz/gojsontoyaml

REPO?=quay.io/coreos/prometheus-operator
TAG?=$(shell git rev-parse --short HEAD)
NAMESPACE?=po-e2e-$(shell LC_ALL=C tr -dc a-z0-9 < /dev/urandom | head -c 13 ; echo '')
KUBECONFIG?=$(HOME)/.kube/config

PROMU := $(GOPATH)/bin/promu
PREFIX ?= $(shell pwd)
ifeq ($(GOBIN),)
GOBIN :=${GOPATH}/bin
endif
pkgs = $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

all: check-license format build test

build: promu
	@$(PROMU) build --prefix $(PREFIX)

short-build:
	go install github.com/coreos/prometheus-operator/cmd/operator

po-crdgen:
	go install github.com/coreos/prometheus-operator/cmd/po-crdgen

crossbuild: promu
	@$(PROMU) crossbuild

test:
	@go test -short $(pkgs)

format:
	go fmt $(pkgs)

check-license:
	./scripts/check_license.sh

container:
	docker build -t $(REPO):$(TAG) .

e2e-test:
	go test -timeout 55m -v ./test/e2e/ $(TEST_RUN_ARGS) --kubeconfig=$(KUBECONFIG) --operator-image=$(REPO):$(TAG) --namespace=$(NAMESPACE)

e2e-status:
	kubectl get prometheus,alertmanager,servicemonitor,statefulsets,deploy,svc,endpoints,pods,cm,secrets,replicationcontrollers --all-namespaces

e2e:
	$(MAKE) container
	$(MAKE) e2e-test

e2e-helm:
	./helm/hack/e2e-test.sh
	# package the chart and verify if they have the version bumped  
	helm/hack/helm-package.sh "alertmanager grafana prometheus prometheus-operator exporter-kube-dns exporter-kube-scheduler exporter-kubelets exporter-node exporter-kube-controller-manager exporter-kube-etcd exporter-kube-state exporter-kubernetes exporter-coredns"
	helm/hack/sync-repo.sh false

clean-e2e:
	kubectl -n $(NAMESPACE) delete prometheus,alertmanager,servicemonitor,statefulsets,deploy,svc,endpoints,pods,cm,secrets,replicationcontrollers --all
	kubectl delete namespace $(NAMESPACE)

promu:
	@go get -u github.com/prometheus/promu

embedmd:
	@go get github.com/campoy/embedmd

po-docgen:
	@go install github.com/coreos/prometheus-operator/cmd/po-docgen

docs: embedmd po-docgen
	$(GOPATH)/bin/embedmd -w `find Documentation contrib/kube-prometheus/README.md -name "*.md"`
	$(GOPATH)/bin/po-docgen api pkg/client/monitoring/v1/types.go > Documentation/api.md
	$(GOPATH)/bin/po-docgen compatibility > Documentation/compatibility.md

generate: jsonnet-docker
	docker run --rm -u=$(shell id -u $(USER)):$(shell id -g $(USER)) -v `pwd`:/go/src/github.com/coreos/prometheus-operator po-jsonnet make generate-deepcopy generate-openapi generate-crd jsonnet generate-bundle generate-kube-prometheus docs


$(GOBIN)/openapi-gen:
	go get -u -v -d k8s.io/code-generator/cmd/openapi-gen
	cd $(GOPATH)/src/k8s.io/code-generator; git checkout release-1.10
	go install k8s.io/code-generator/cmd/openapi-gen

$(GOBIN)/deepcopy-gen:
	go get -u -v -d k8s.io/code-generator/cmd/deepcopy-gen
	cd $(GOPATH)/src/k8s.io/code-generator; git checkout release-1.10
	go install k8s.io/code-generator/cmd/deepcopy-gen

openapi-gen: $(GOBIN)/openapi-gen

deepcopy-gen: $(GOBIN)/deepcopy-gen

generate-deepcopy: deepcopy-gen
	$(GOBIN)/deepcopy-gen -i github.com/coreos/prometheus-operator/pkg/client/monitoring/v1 --go-header-file="$(GOPATH)/src/github.com/coreos/prometheus-operator/.header" -v=4 --logtostderr --bounding-dirs "github.com/coreos/prometheus-operator/pkg/client" --output-file-base zz_generated.deepcopy
	$(GOBIN)/deepcopy-gen -i github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1 --go-header-file="$(GOPATH)/src/github.com/coreos/prometheus-operator/.header" -v=4 --logtostderr --bounding-dirs "github.com/coreos/prometheus-operator/pkg/client" --output-file-base zz_generated.deepcopy

generate-openapi: openapi-gen
	$(GOBIN)/openapi-gen  -i github.com/coreos/prometheus-operator/pkg/client/monitoring/v1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/api/core/v1  -p github.com/coreos/prometheus-operator/pkg/client/monitoring/v1 --go-header-file="$(GOPATH)/src/github.com/coreos/prometheus-operator/.header"

generate-bundle:
	hack/generate-bundle.sh

generate-kube-prometheus:
	# Update the Prometheus Operator version in kube-prometheus
	sed -i                                                            \
		"s/prometheusOperator: 'v.*',/prometheusOperator: 'v$(shell cat VERSION)',/" \
		contrib/kube-prometheus/jsonnet/kube-prometheus/prometheus-operator/prometheus-operator.libsonnet;
	cd contrib/kube-prometheus; $(MAKE) generate-raw

jsonnet: jb
	cd hack/generate; jb install
	jsonnet -J hack/generate/vendor hack/generate/prometheus-operator.jsonnet | gojsontoyaml > example/non-rbac/prometheus-operator.yaml
	jsonnet -J hack/generate/vendor hack/generate/prometheus-operator-rbac.jsonnet | gojsontoyaml > example/rbac/prometheus-operator/prometheus-operator.yaml

jb:
	go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb

jsonnet-docker:
	docker build -f scripts/jsonnet/Dockerfile -t po-jsonnet .

helm-sync-s3:
	helm/hack/helm-package.sh "alertmanager grafana prometheus prometheus-operator exporter-kube-dns exporter-kube-scheduler exporter-kubelets exporter-node exporter-kube-controller-manager exporter-kube-etcd exporter-kube-state exporter-kubernetes exporter-coredns"
	helm/hack/sync-repo.sh true
	helm/hack/helm-package.sh kube-prometheus
	helm/hack/sync-repo.sh true

generate-crd: generate-openapi po-crdgen
	po-crdgen prometheus > example/prometheus-operator-crd/prometheus.crd.yaml
	po-crdgen alertmanager > example/prometheus-operator-crd/alertmanager.crd.yaml
	po-crdgen servicemonitor > example/prometheus-operator-crd/servicemonitor.crd.yaml

.PHONY: all build crossbuild test format check-license container e2e-test e2e-status e2e clean-e2e embedmd apidocgen docs generate-crd jb

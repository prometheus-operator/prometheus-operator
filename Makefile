REPO?=quay.io/coreos/prometheus-operator
TAG?=$(shell git rev-parse --short HEAD)
NAMESPACE?=po-e2e-$(shell LC_CTYPE=C tr -dc a-z0-9 < /dev/urandom | head -c 13 ; echo '')
KUBECONFIG?=$(HOME)/.kube/config

PROMU := $(GOPATH)/bin/promu
PREFIX ?= $(shell pwd)

pkgs = $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

all: check-license format build test

build: promu
	@$(PROMU) build --prefix $(PREFIX)

short-build:
	go install github.com/coreos/prometheus-operator/cmd/operator

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
	$(GOPATH)/bin/embedmd -w `find Documentation -name "*.md"`
	$(GOPATH)/bin/po-docgen api pkg/client/monitoring/v1/types.go > Documentation/api.md
	$(GOPATH)/bin/po-docgen compatibility > Documentation/compatibility.md

generate: jsonnet-docker
	docker run --rm -v `pwd`:/go/src/github.com/coreos/prometheus-operator po-jsonnet make generate-deepcopy jsonnet generate-bundle docs generate-kube-prometheus

deepcopy-gen:
	go get -u -v -d k8s.io/code-generator/cmd/deepcopy-gen
	cd $(GOPATH)/src/k8s.io/code-generator; git checkout release-1.8
	go install k8s.io/code-generator/cmd/deepcopy-gen

generate-deepcopy: deepcopy-gen
	deepcopy-gen -i github.com/coreos/prometheus-operator/pkg/client/monitoring/v1 --go-header-file="$(GOPATH)/src/github.com/coreos/prometheus-operator/.header" -v=4 --logtostderr --bounding-dirs "github.com/coreos/prometheus-operator/pkg/client" --output-file-base zz_generated.deepcopy
	deepcopy-gen -i github.com/coreos/prometheus-operator/pkg/client/monitoring/v1alpha1 --go-header-file="$(GOPATH)/src/github.com/coreos/prometheus-operator/.header" -v=4 --logtostderr --bounding-dirs "github.com/coreos/prometheus-operator/pkg/client" --output-file-base zz_generated.deepcopy

generate-bundle:
	hack/generate-bundle.sh

generate-kube-prometheus:
	cd contrib/kube-prometheus; $(MAKE) generate-raw

jsonnet:
	jsonnet -J /ksonnet-lib hack/generate/prometheus-operator.jsonnet | json2yaml > example/non-rbac/prometheus-operator.yaml
	jsonnet -J /ksonnet-lib hack/generate/prometheus-operator-rbac.jsonnet | json2yaml > example/rbac/prometheus-operator/prometheus-operator.yaml
	jsonnet -J /ksonnet-lib hack/generate/prometheus-operator-rbac.jsonnet | json2yaml > contrib/kube-prometheus/manifests/prometheus-operator/prometheus-operator.yaml

jsonnet-docker:
	docker build -f scripts/jsonnet/Dockerfile -t po-jsonnet .

helm-sync-s3:
	helm/hack/helm-package.sh "alertmanager grafana prometheus prometheus-operator exporter-kube-dns exporter-kube-scheduler exporter-kubelets exporter-node exporter-kube-controller-manager exporter-kube-etcd exporter-kube-state exporter-kubernetes"
	helm/hack/sync-repo.sh
	helm/hack/helm-package.sh kube-prometheus
	helm/hack/sync-repo.sh

.PHONY: all build crossbuild test format check-license container e2e-test e2e-status e2e clean-e2e embedmd apidocgen docs

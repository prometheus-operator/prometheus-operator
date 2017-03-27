REPO?=quay.io/coreos/prometheus-operator
TAG?=$(shell git rev-parse --short HEAD)
NAMESPACE?=prometheus-operator-e2e-tests-$(shell LC_CTYPE=C tr -dc a-z0-9 < /dev/urandom | head -c 13 ; echo '')

PROMU := $(GOPATH)/bin/promu
PREFIX ?= $(shell pwd)

CLUSTER_IP?=$(shell kubectl config view --minify | grep server: | cut -f 3 -d ":" | tr -d "//")

pkgs = $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

all: check-license format build test

build: promu
	@$(PROMU) build --prefix $(PREFIX)

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
	go test -timeout 20m -v ./test/e2e/ $(TEST_RUN_ARGS) --kubeconfig "$(HOME)/.kube/config" --operator-image=$(REPO):$(TAG) --namespace=$(NAMESPACE) --cluster-ip=$(CLUSTER_IP)

e2e-status:
	kubectl get prometheus,alertmanager,servicemonitor,statefulsets,deploy,svc,endpoints,pods,cm,secrets,replicationcontrollers --all-namespaces

e2e:
	$(MAKE) container
	$(MAKE) e2e-test

clean-e2e:
	kubectl -n $(NAMESPACE) delete prometheus,alertmanager,servicemonitor,statefulsets,deploy,svc,endpoints,pods,cm,secrets,replicationcontrollers --all
	kubectl delete namespace $(NAMESPACE)

promu:
	@go get -u github.com/prometheus/promu

embedmd:
	@go get github.com/campoy/embedmd

apidocgen:
	@go install github.com/coreos/prometheus-operator/cmd/apidocgen

docs: embedmd apidocgen
	embedmd -w `find Documentation -name "*.md"`
	apidocgen pkg/client/monitoring/v1alpha1/types.go > Documentation/api.md

generate:
	hack/generate.sh
	@$(MAKE) docs

.PHONY: all build crossbuild test format check-license container e2e-test e2e-status e2e clean-e2e embedmd apidocgen docs

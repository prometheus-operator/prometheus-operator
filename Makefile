REPO?=quay.io/coreos/prometheus-operator
TAG?=$(shell git rev-parse --short HEAD)
NAMESPACE?=prometheus-operator-e2e-tests-$(shell LC_CTYPE=C tr -dc a-z0-9 < /dev/urandom | head -c 13 ; echo '')

CLUSTER_IP?=$(shell minikube ip)

pkgs = $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

all: check-license format build test

build:
	go build github.com/coreos/prometheus-operator/cmd/operator

test:
	@go test -short $(pkgs)

format:
	go fmt $(pkgs)

check-license:
	./scripts/check_license.sh

container:
	GOOS=linux $(MAKE) build
	docker build -t $(REPO):$(TAG) .

e2e-test:
	go test -timeout 20m -v ./test/e2e/ $(TEST_RUN_ARGS) --kubeconfig "$(HOME)/.kube/config" --operator-image=quay.io/coreos/prometheus-operator:$(TAG) --namespace=$(NAMESPACE) --cluster-ip=$(CLUSTER_IP)

e2e-status:
	kubectl get prometheus,alertmanager,servicemonitor,statefulsets,deploy,svc,endpoints,pods,cm --all-namespaces

e2e:
	$(MAKE) container
	$(MAKE) e2e-test

clean-e2e:
	kubectl -n $(NAMESPACE) delete prometheus,alertmanager,servicemonitor,statefulsets,deploy,svc,endpoints,pods,cm --all
	kubectl delete namespace $(NAMESPACE)

.PHONY: all build test format check-license container e2e-test e2e-status e2e clean-e2e

all: build

REPO?=quay.io/coreos/prometheus-operator
TAG?=$(shell git rev-parse --short HEAD)
NAMESPACE?=prometheus-operator-e2e-tests-$(shell head /dev/urandom | tr -dc a-z0-9 | head -c 13 ; echo '')

build:
	./scripts/check_license.sh
	go build github.com/coreos/prometheus-operator/cmd/operator

container:
	GOOS=linux $(MAKE) build
	docker build -t $(REPO):$(TAG) .

e2e-test:
	go test -v ./test/e2e/ --kubeconfig "$(HOME)/.kube/config" --operator-image=quay.io/coreos/prometheus-operator:$(TAG) --namespace=$(NAMESPACE)

e2e:
	$(MAKE) container
	$(MAKE) e2e-test

clean-e2e:
	kubectl delete namespace prometheus-operator-e2e-tests

.PHONY: all build container e2e clean-e2e

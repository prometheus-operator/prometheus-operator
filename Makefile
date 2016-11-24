all: build

REPO = quay.io/coreos/prometheus-operator
TAG = latest

build:
	./scripts/check_license.sh
	go build github.com/coreos/prometheus-operator/cmd/operator

container:
	GOOS=linux $(MAKE) build
	docker build -t $(REPO):$(TAG) .

e2e:
	go test -v ./test/e2e/ --kubeconfig "$(HOME)/.kube/config" --operator-image=quay.io/coreos/prometheus-operator

clean-e2e:
	kubectl delete namespace prometheus-operator-e2e-tests

.PHONY: all build container e2e clean-e2e

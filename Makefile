all: build

REPO = quay.io/coreos/prometheus-operator
TAG = latest

build:
	./scripts/check_license.sh
	go build github.com/coreos/prometheus-operator/cmd/operator

container:
	GOOS=linux $(MAKE) build
	docker build -t $(REPO):$(TAG) .

.PHONY: all build container

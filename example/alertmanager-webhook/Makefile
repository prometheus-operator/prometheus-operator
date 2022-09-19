IMAGE ?= quay.io/prometheus-operator/prometheus-alertmanager-test-webhook
TAG ?= latest

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags netgo -o linux/amd64/ .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags netgo -o linux/arm64/ .

.PHONY: push
push: build
	buildah manifest create $(IMAGE):$(TAG)
	buildah bud --arch amd64 --build-arg ARCH=amd64 --manifest $(IMAGE):$(TAG) -t $(IMAGE):$(TAG)-amd64 .
	buildah bud --arch arm64 --build-arg ARCH=arm64 --manifest $(IMAGE):$(TAG) -t $(IMAGE):$(TAG)-arm64 .
	buildah manifest push --all $(IMAGE):$(TAG) docker://$(IMAGE):$(TAG)

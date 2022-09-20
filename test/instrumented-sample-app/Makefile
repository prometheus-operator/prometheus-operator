IMAGE ?= quay.io/prometheus-operator/instrumented-sample-app
TAG ?= latest
VERSION=$(shell cat VERSION)

.PHONY: build
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags netgo -o linux/amd64/ .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -tags netgo -o linux/arm64/ .

.PHONY: push
push: build
	buildah manifest create $(IMAGE):$(TAG)
	buildah bud --arch amd64 --build-arg ARCH=amd64 --build-arg VERSION=$(VERSION) --manifest $(IMAGE):$(TAG) -t $(IMAGE):$(TAG)-amd64 .
	buildah bud --arch arm64 --build-arg ARCH=arm64 --build-arg VERSION=$(VERSION) --manifest $(IMAGE):$(TAG) -t $(IMAGE):$(TAG)-arm64 .
	buildah manifest push --all $(IMAGE):$(TAG) docker://$(IMAGE):$(TAG)

generate-certs:
	mkdir -p certs && \
	openssl req -newkey rsa:2048 \
		-new -nodes -x509 \
		-days 3650 \
		-out certs/cert.pem \
		-keyout certs/key.pem \
		-subj "/C=US/ST=California/L=Mountain View/O=Your Organization/OU=Your Unit/CN=localhost"

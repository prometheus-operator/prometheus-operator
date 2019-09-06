REG := quay.io/coreos
APP := instrumented-sample-app
VERSION ?= $(shell cat VERSION)
GO111MODULE?=on
GOPROXY?=http://proxy.golang.org
export GO111MODULE
export GOPROXY

all: container push

build:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -o instrumented-sample-app main.go

container:
	@docker build -t $(REG)/$(APP):$(VERSION) --file Dockerfile ../../

push:
	@docker push $(REG)/$(APP):$(VERSION)

generate-certs:
	mkdir -p certs && \
	openssl req -newkey rsa:2048 \
		-new -nodes -x509 \
		-days 3650 \
		-out certs/cert.pem \
		-keyout certs/key.pem \
		-subj "/C=US/ST=California/L=Mountain View/O=Your Organization/OU=Your Unit/CN=localhost"

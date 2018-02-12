TAG?=$(shell git rev-parse --short HEAD)
PREFIX ?= $(shell pwd)
REPO = github.com/ant31/crd-validation
pkgs = $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

all: check-license format build test

run:
	go run cmd/openapi-crd-gen/main.go

install: openapi-gen


test:
	@go test -short $(pkgs)

format:
	go fmt $(pkgs)

openapi-gen:
	go get -u -v -d k8s.io/code-generator/cmd/openapi-gen
	cd $(GOPATH)/src/k8s.io/code-generator; git checkout release-1.8
	go install k8s.io/code-generator/cmd/openapi-gen


.PHONY: all build test format generate-openapi openapi-gen

REG := quay.io/coreos
APP := basic-auth-test-app
VERSION ?= $(shell cat VERSION)

all: build push 

build: 
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o basic-auth-test-app main.go
	@docker build --build-arg VERSION=$(VERSION) -t $(REG)/$(APP):$(VERSION) --file $(APP).dockerfile .
	@rm $(APP)

push:
	@docker push $(REG)/$(APP):$(VERSION)

REG := quay.io/coreos
APP := basic-auth-test-app
VERSION ?= $(shell cat VERSION)

all: build push 

build: 
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -o basic-auth-test-app main.go
	@docker build --build-arg VERSION=$(VERSION) -t $(REG)/$(APP):$(VERSION) --file $(APP).dockerfile .
	@rm $(APP)

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

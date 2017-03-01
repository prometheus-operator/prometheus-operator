all: build

FLAGS =
ENVVAR = GOOS=linux GOARCH=amd64 CGO_ENABLED=0
REGISTRY = quay.io/coreos
TAG = latest

build:
	$(ENVVAR) go build -o grafana-watcher main.go

image: build
	docker build -t ${REGISTRY}/grafana-watcher:$(TAG) .

push: image
	docker push ${REGISTRY}/grafana-watcher:$(TAG)

clean:
	rm -f grafana-watcher

.PHONY: all build image push clean

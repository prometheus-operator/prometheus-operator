.PHONY: image

image:
	docker build -f ../../scripts/jsonnet/Dockerfile -t po-jsonnet ../../

generate: image
	@echo ">> Compiling assets and generating Kubernetes manifests"
	docker run --rm -u=$(shell id -u $(USER)):$(shell id -g $(USER)) -v `pwd`:/go/src/github.com/coreos/prometheus-operator/contrib/kube-prometheus --workdir /go/src/github.com/coreos/prometheus-operator/contrib/kube-prometheus po-jsonnet make generate-raw

generate-raw:
	./hack/scripts/build-jsonnet.sh example-dist/base/kube-prometheus.jsonnet manifests

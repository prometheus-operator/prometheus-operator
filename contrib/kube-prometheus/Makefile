JSONNET_FMT := jsonnet fmt -n 2 --max-blank-lines 2 --string-style s --comment-style s

image:
	docker build -f ../../scripts/jsonnet/Dockerfile -t po-jsonnet ../../

generate: image
	@echo ">> Compiling assets and generating Kubernetes manifests"
	docker run --rm -u=$(shell id -u $(USER)):$(shell id -g $(USER)) -v $(shell dirname $(dir $(abspath $(dir $$PWD)))):/go/src/github.com/coreos/prometheus-operator/ --workdir /go/src/github.com/coreos/prometheus-operator/contrib/kube-prometheus po-jsonnet make generate-raw

crdtojsonnet:
	cat ../../example/prometheus-operator-crd/alertmanager.crd.yaml | gojsontoyaml -yamltojson > jsonnet/kube-prometheus/prometheus-operator/alertmanager-crd.libsonnet
	cat ../../example/prometheus-operator-crd/prometheus.crd.yaml | gojsontoyaml -yamltojson > jsonnet/kube-prometheus/prometheus-operator/prometheus-crd.libsonnet
	cat ../../example/prometheus-operator-crd/servicemonitor.crd.yaml | gojsontoyaml -yamltojson > jsonnet/kube-prometheus/prometheus-operator/servicemonitor-crd.libsonnet
	cat ../../example/prometheus-operator-crd/rulefile.crd.yaml | gojsontoyaml -yamltojson > jsonnet/kube-prometheus/prometheus-operator/rulefile-crd.libsonnet

generate-raw: crdtojsonnet fmt
	jb install
	./build.sh

fmt:
	find . -name 'vendor' -prune -o -name '*.libsonnet' -o -name '*.jsonnet' -print | \
		xargs -n 1 -- $(JSONNET_FMT) -i

test: image
	@echo ">> Compiling assets and generating Kubernetes manifests"
	docker run --rm -u=$(shell id -u $(USER)):$(shell id -g $(USER)) -v $(shell dirname $(dir $(abspath $(dir $$PWD)))):/go/src/github.com/coreos/prometheus-operator/ --workdir /go/src/github.com/coreos/prometheus-operator/contrib/kube-prometheus po-jsonnet make test-raw

test-raw: crdtojsonnet
	jb install
	./test.sh

.PHONY: image generate crdtojsonnet generate-raw test test-raw fmt

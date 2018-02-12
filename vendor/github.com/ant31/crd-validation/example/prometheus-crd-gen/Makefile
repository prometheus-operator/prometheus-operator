OPENAPI_GEN := $(shell command -v openapi-gen 2> /dev/null)
REPO :=github.com/ant31/crd-validation/example/prometheus-crd-gen
ifeq ($(GOBIN),)
GOBIN :=${GOPATH}/bin
endif
DATE := $(shell date '+%Y-%m-%d %H:%M:%S')

all: generate-crds

$(GOBIN)/openapi-gen:
	go get -u -v -d k8s.io/code-generator/cmd/openapi-gen
	cd $(GOPATH)/src/k8s.io/code-generator; git checkout release-1.8
	go install k8s.io/code-generator/cmd/openapi-gen



v1/openapi_generated.go: $(GOBIN)/openapi-gen
	openapi-gen  -i $(REPO)/v1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/api/core/v1  -p $(REPO)/v1 --go-header-file="$(GOPATH)/src/github.com/ant31/crd-validation/.header"

generate-openapi: v1/openapi_generated.go

build: generate-openapi
	go install $(REPO)

$(GOBIN)/prometheus-crd-gen: build

install: $(GOBIN)/prometheus-crd-gen


generated-crd/prometheus.crd.yaml: $(GOBIN)/prometheus-crd-gen
	prometheus-crd-gen --kind=Prometheus --plural=prometheuses --apigroup=monitoring.coreos.com --labels generation-date="$(DATE)" --scope=Namespaced --version=v1 --spec-name=github.com/ant31/crd-validation/example/prometheus-crd-gen/v1.Prometheus > generated-crd/prometheus.crd.yaml

generated-crd/alertmanager.crd.yaml: $(GOBIN)/prometheus-crd-gen
	prometheus-crd-gen --kind=Alertmanager --plural=alertmanagers --apigroup=monitoring.coreos.com --labels generation-date="$(DATE)" --scope=Namespaced --version=v1 --spec-name=github.com/ant31/crd-validation/example/prometheus-crd-gen/v1.Alertmanager > generated-crd/alertmanager.crd.yaml

generated-crd/servicemonitor.crd.yaml: $(GOBIN)/prometheus-crd-gen
	prometheus-crd-gen --kind=ServiceMonitor --plural=servicemonitors --apigroup=monitoring.coreos.com --labels generation-date="$(DATE)" --scope=Cluster --version=v1 --spec-name=github.com/ant31/crd-validation/example/prometheus-crd-gen/v1.ServiceMonitor > generated-crd/servicemonitor.crd.yaml

generate-crds: clean generated-crd/prometheus.crd.yaml generated-crd/alertmanager.crd.yaml generated-crd/servicemonitor.crd.yaml

.PHONY: openapi-gen build all


clean:
	rm -rf generated-crd/*.yaml

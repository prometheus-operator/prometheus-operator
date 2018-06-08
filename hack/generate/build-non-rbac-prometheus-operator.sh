#!/usr/bin/env bash

jsonnet -J hack/generate/vendor hack/generate/prometheus-operator-non-rbac.jsonnet | gojsontoyaml > example/non-rbac/prometheus-operator.yaml

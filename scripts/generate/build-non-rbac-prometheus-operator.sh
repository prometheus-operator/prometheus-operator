#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

jsonnet -J scripts/generate/vendor scripts/generate/prometheus-operator-non-rbac.jsonnet | gojsontoyaml > example/non-rbac/prometheus-operator.yaml

#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

hack/concat-kubernetes-manifests.sh example/rbac/prometheus-operator/*.yaml > bundle.yaml

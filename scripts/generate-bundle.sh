#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

function concat() {
    awk 'FNR==1{print "---"}1' "$@" | awk '{if (NR!=1) {print}}'
}

# shellcheck disable=SC2046
concat $(find example/rbac/prometheus-operator -name '*.yaml' | sort | grep -v service-monitor) > bundle.yaml
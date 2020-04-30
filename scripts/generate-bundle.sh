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

function generate_bundle() {
    # shellcheck disable=SC2046
    concat $(find "$@" -maxdepth 1 -name '*.yaml' | sort | grep -v service-monitor)
}

generate_bundle example/rbac/prometheus-operator example/prometheus-operator-crd > bundle.yaml

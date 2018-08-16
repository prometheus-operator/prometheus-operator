#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

kubectl get pods --all-namespaces -l app="${1}" -ojsonpath="{.items[*].spec.containers[?(@.name==\"$1\")].image}"

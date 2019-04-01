#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")

"${SCRIPT_DIR}"/../../../../scripts/create-minikube.sh

(
    cd "${SCRIPT_DIR}"/../.. || exit
    kubectl apply -f manifests
    KUBECONFIG=~/.kube/config make test-e2e
)

"${SCRIPT_DIR}"/../../../../scripts/delete-minikube.sh

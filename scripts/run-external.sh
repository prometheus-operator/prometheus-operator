#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

if [[ -z "$1" ]]; then
	echo "missing cluster name"
	exit 1
fi

apiserver=$(kubectl config view -o jsonpath="{.clusters[?(@.name == \"$1\")].cluster.server}")
cafile=$(kubectl config view -o jsonpath="{.clusters[?(@.name == \"$1\")].cluster.certificate-authority}")
certfile=$(kubectl config view -o jsonpath="{.users[?(@.name == \"$1\")].user.client-certificate}")
keyfile=$(kubectl config view -o jsonpath="{.users[?(@.name == \"$1\")].user.client-key}")

./operator --apiserver="${apiserver}" --ca-file="${cafile}" --cert-file="${certfile}" --key-file="${keyfile}"

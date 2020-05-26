#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

PO=$(jsonnet -J scripts/generate/vendor scripts/generate/prometheus-operator-rbac.jsonnet)
echo "$PO" | jq -r 'keys[]' | while read -r file
do
    echo "$PO" | jq -r ".[\"${file}\"]" | gojsontoyaml > "example/rbac/prometheus-operator/${file}"
done

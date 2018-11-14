#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

rm -rf tmp
mkdir tmp
jsonnet -J hack/generate/vendor hack/generate/prometheus-operator-rbac.jsonnet > tmp/po.json
jq -r 'keys[]' tmp/po.json | while read -r file
do
    jq -r ".[\"${file}\"]" tmp/po.json | gojsontoyaml > "example/rbac/prometheus-operator/${file}"
done

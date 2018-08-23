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
mapfile -t files < <(jq -r 'keys[]' tmp/po.json)
for file in "${files[@]}"
do
    jq -r ".[\"${file}\"]" tmp/po.json | gojsontoyaml > "example/rbac/prometheus-operator/${file}"
done

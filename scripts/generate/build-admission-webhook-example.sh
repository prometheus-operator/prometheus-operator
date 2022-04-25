#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

AW=$(jsonnet -J scripts/generate/vendor scripts/generate/admission-webhook.jsonnet)
echo "$AW" | jq -r 'keys[]' | while read -r file
do
    echo "$AW" | jq -r ".[\"${file}\"]" | gojsontoyaml > "example/admission-webhook/${file}"
done

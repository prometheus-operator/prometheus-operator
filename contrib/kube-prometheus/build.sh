#!/usr/bin/env bash
set -e
set -x
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

                                               # optional, but we would like to generate yaml, not json
jsonnet -J vendor -m manifests ${1-example.jsonnet} | xargs -I{} sh -c 'cat $1 | gojsontoyaml > $1.yaml; rm -f $1' -- {}


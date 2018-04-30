#!/usr/bin/env bash
set -e
set -x

                                               # optional, but we would like to generate yaml, not json
jsonnet -J vendor -m manifests example.jsonnet | xargs -I{} sh -c 'cat $1 | gojsontoyaml > $1.yaml; rm $1' -- {}


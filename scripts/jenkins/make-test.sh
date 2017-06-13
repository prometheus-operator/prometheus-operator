#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

docker run \
       --rm \
       -v $PWD:/go/src/github.com/coreos/prometheus-operator \
       -w /go/src/github.com/coreos/prometheus-operator \
       golang make test

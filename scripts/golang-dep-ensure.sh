#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

dep ensure


git diff --exit-code

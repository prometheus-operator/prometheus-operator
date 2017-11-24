#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

go get -v github.com/kardianos/govendor

cd vendor
rm -r $(ls -I "vendor.json" )
cd ..

govendor sync

git diff --exit-code

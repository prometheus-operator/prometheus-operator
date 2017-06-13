#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

PO_QUAY_REPO=quay.io/coreos/prometheus-operator-dev

docker login -u="$QUAY_ROBOT_USERNAME" -p="$QUAY_ROBOT_SECRET" quay.io

docker tag \
       $PO_QUAY_REPO:$BUILD_ID \
       $PO_QUAY_REPO:master

docker push $PO_QUAY_REPO:master

docker logout quay.io


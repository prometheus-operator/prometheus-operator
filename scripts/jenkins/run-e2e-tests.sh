#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

DOCKER_SOCKET=/var/run/docker.sock
PO_QUAY_REPO=quay.io/coreos/prometheus-operator-dev

docker build -t cluster-setup-env scripts/jenkins/.
docker run \
       --rm \
       -v $PWD:$PWD -v $DOCKER_SOCKET:$DOCKER_SOCKET \
       cluster-setup-env \
       /bin/bash -c "cd $PWD && make crossbuild"

docker build -t $PO_QUAY_REPO:$BUILD_ID .
docker login -u="$QUAY_ROBOT_USERNAME" -p="$QUAY_ROBOT_SECRET" quay.io
docker push $PO_QUAY_REPO:$BUILD_ID

docker run \
       --rm \
       -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY \
       -e REPO=$PO_QUAY_REPO -e TAG=$BUILD_ID \
       -v $PWD:/go/src/github.com/coreos/prometheus-operator \
       -w /go/src/github.com/coreos/prometheus-operator/scripts/jenkins \
       cluster-setup-env \
       /bin/bash -c "make"

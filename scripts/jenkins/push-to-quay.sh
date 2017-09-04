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


# Retry pushing docker image multiple times to prevent net/http: TLS handshake
# timeout

retry=0
maxRetries=5
until [ ${retry} -ge ${maxRetries} ]
do
    docker push $PO_QUAY_REPO:master && break
    retry=$[${retry}+1]
done

docker logout quay.io

if [ ${retry} -ge ${maxRetries} ]; then
    echo "Failed to push docker image after ${maxRetries} attempts!"
    exit 1
fi



#!/usr/bin/env bash
# This is a cleanup script, if one command fails we still want all others to run
# set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x

export {TF_GET_OPTIONS,TF_PLAN_OPTIONS,TF_APPLY_OPTIONS,TF_DESTROY_OPTIONS}="-no-color"

CLUSTER="po-$(git rev-parse --short HEAD)-${BUILD_ID}"
TF_VAR_tectonic_cluster_name="${CLUSTER}"
TF_VAR_tectonic_dns_name="${CLUSTER}"
TECTONIC_INSTALLER_DIR=/go/src/github.com/coreos/tectonic-installer

docker run \
       --rm \
       -v $PWD/build/:$TECTONIC_INSTALLER_DIR/build/ \
       -v ~/.ssh:$HOME/.ssh \
       -e AWS_ACCESS_KEY_ID \
       -e AWS_SECRET_ACCESS_KEY \
       -e TF_GET_OPTIONS \
       -e TF_DESTROY_OPTIONS \
       -e CLUSTER=${CLUSTER} \
       -w $TECTONIC_INSTALLER_DIR \
       -e TF_VAR_tectonic_cluster_name=${TF_VAR_tectonic_cluster_name} \
       -e TF_VAR_tectonic_dns_name=${TF_VAR_tectonic_dns_name} \
       quay.io/coreos/tectonic-installer:master \
       /bin/bash -c "make destroy || make destroy || make destroy"

docker rmi quay.io/coreos/prometheus-operator-dev:$BUILD_ID

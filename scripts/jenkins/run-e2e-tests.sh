#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u
# print each command before executing it
set -x


# Push docker image

DOCKER_SOCKET=/var/run/docker.sock
PO_QUAY_REPO=quay.io/coreos/prometheus-operator-dev

docker build -t docker-golang-env -f scripts/jenkins/docker-golang-env/Dockerfile .

docker run \
       --rm \
       -v $PWD:$PWD -v $DOCKER_SOCKET:$DOCKER_SOCKET \
       docker-golang-env \
       /bin/bash -c "cd $PWD && make crossbuild"

docker build -t $PO_QUAY_REPO:$BUILD_ID .
docker login -u="$QUAY_ROBOT_USERNAME" -p="$QUAY_ROBOT_SECRET" quay.io
docker push $PO_QUAY_REPO:$BUILD_ID


# Bring up k8s cluster

export {TF_GET_OPTIONS,TF_PLAN_OPTIONS,TF_APPLY_OPTIONS,TF_DESTROY_OPTIONS}="-no-color"

CLUSTER="po-$(git rev-parse --short HEAD)-${BUILD_ID}"
TF_VAR_tectonic_cluster_name="${CLUSTER}"
TF_VAR_tectonic_dns_name="${CLUSTER}"
TECTONIC_INSTALLER_DIR=/go/src/github.com/coreos/tectonic-installer
PO_DIR=/go/src/github.com/coreos/prometheus-operator
KUBECONFIG="${PO_DIR}/build/${CLUSTER}/generated/auth/kubeconfig"
TECTONIC_INSTALLER="quay.io/coreos/tectonic-installer:master"

mkdir -p build/${CLUSTER}
cp ${WORKSPACE}/scripts/jenkins/kubernetes-vanilla.tfvars build/${CLUSTER}/terraform.tfvars

docker pull $TECTONIC_INSTALLER
docker run \
       --rm \
       -v $PWD/build/:$TECTONIC_INSTALLER_DIR/build/ \
       -v ~/.ssh:$HOME/.ssh \
       -e AWS_ACCESS_KEY_ID \
       -e AWS_SECRET_ACCESS_KEY \
       -e TF_GET_OPTIONS \
       -e TF_PLAN_OPTIONS \
       -e TF_APPLY_OPTIONS \
       -e CLUSTER=${CLUSTER} \
       -e TF_VAR_tectonic_cluster_name=${TF_VAR_tectonic_cluster_name} \
       -e TF_VAR_tectonic_dns_name=${TF_VAR_tectonic_dns_name} \
       -w $TECTONIC_INSTALLER_DIR \
       $TECTONIC_INSTALLER \
       /bin/bash -c "touch license secret && make plan && make apply"

docker build \
       -t kubectl-env \
       -f scripts/jenkins/kubectl-env/Dockerfile \
       .

sleep 5m
docker run \
       --rm \
       -v $PWD:$PO_DIR \
       -w $PO_DIR \
       -e KUBECONFIG=${KUBECONFIG} \
       kubectl-env \
       /bin/bash -c "timeout 900 ./scripts/jenkins/wait-for-cluster.sh 4"


# Run e2e tests

docker run \
       --rm \
       -v $PWD:$PO_DIR \
       -w $PO_DIR \
       -e KUBECONFIG=${KUBECONFIG} \
       -e REPO=$PO_QUAY_REPO \
       -e TAG=$BUILD_ID \
       kubectl-env \
       /bin/bash -c "make e2e-test"


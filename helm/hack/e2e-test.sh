#!/usr/bin/env bash
#// TODO move this test and dependencies to a docker container
set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# Install and initialize helm/tiller
HELM_URL=https://storage.googleapis.com/kubernetes-helm
HELM_TARBALL=helm-v2.7.2-linux-amd64.tar.gz
NAMESPACE="helm-monitoring"
CUR_DIR=$(dirname "${BASH_SOURCE[0]}")
wget -q ${HELM_URL}/${HELM_TARBALL}
tar xzfv ${HELM_TARBALL}

# # Clean up tarball
rm -f ${HELM_TARBALL}
sudo mv linux-amd64/helm /usr/local/bin
# setup tiller
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller --upgrade

# wait until all minkube pods, including tiller, are in reading state
"${CUR_DIR}"/wait-pods-running-state.sh kube-system 

kubectl create ns ${NAMESPACE}

# replace current http repository to the helm path
sed -ie 's/    repository/#    repository/g' "$(pwd)"/helm/*/requirements.yaml
sed -ie 's/#e2e-repository/repository/g' "$(pwd)"/helm/*/requirements.yaml

# package charts and install all
"${CUR_DIR}"/helm-package.sh prometheus-operator
"${CUR_DIR}"/helm-package.sh kube-prometheus

helm install --namespace="${NAMESPACE}" "$(pwd)/helm/prometheus-operator" --name prometheus-operator
helm install --namespace="${NAMESPACE}" "$(pwd)/helm/kube-prometheus" --name kube-prometheus

# check if all pods are ready 
"${CUR_DIR}"/wait-pods-running-state.sh ${NAMESPACE}

# reset helm changes
git reset --hard

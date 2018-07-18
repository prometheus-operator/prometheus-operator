#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

# Install and initialize helm/tiller
HELM_URL=https://storage.googleapis.com/kubernetes-helm
HELM_TARBALL=helm-v2.7.2-linux-amd64.tar.gz
HELM_PACKAGES=${1}
HELM_BUCKET_NAME="coreos-charts"
HELM_CHARTS_DIRECTORY=${2:-"$(pwd)/helm"}
HELM_CHARTS_PACKAGED_DIR=${3:-"/tmp/helm-packaged"}
HELM_REPO_URL=https://s3-eu-west-1.amazonaws.com/${HELM_BUCKET_NAME}/stable/
HELM_INDEX="${HELM_CHARTS_PACKAGED_DIR}/index.yaml"

wget ${HELM_URL}/${HELM_TARBALL}
tar xzfv ${HELM_TARBALL}
PATH=${PATH}:$(pwd)/linux-amd64/
export PATH

# Clean up tarball
rm -f ${HELM_TARBALL}

# Package helm and dependencies
mkdir -p "${HELM_CHARTS_PACKAGED_DIR}"
helm init --client-only
helm repo add ${HELM_BUCKET_NAME} ${HELM_REPO_URL} 

# check if charts has dependencies,
for chart in ${HELM_PACKAGES}
do
    (
        # update dependencies before package the chart
        cd "${HELM_CHARTS_DIRECTORY}/${chart}"
        helm dep update
        helm package . -d "${HELM_CHARTS_PACKAGED_DIR}"
    )
done

# donwload the current remote index.yaml 
if [ ! -f "${HELM_INDEX}" ]; then
    wget ${HELM_REPO_URL}index.yaml -O "${HELM_INDEX}"
fi

helm repo index "${HELM_CHARTS_PACKAGED_DIR}" --url "${HELM_REPO_URL}" --debug --merge "${HELM_INDEX}"

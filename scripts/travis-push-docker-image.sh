#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

function defer {
	docker logout quay.io
}
trap defer EXIT

if [[ "${TRAVIS_PULL_REQUEST}" != "false" ]]; then
	exit 0
fi

# Push to Quay '-dev' repo if not a git tag or master branch build
export REPO="quay.io/coreos/prometheus-operator"
export REPO_PROMETHEUS_CONFIG_RELOADER="quay.io/coreos/prometheus-config-reloader"
if [[ "${TRAVIS_TAG}" == "" ]] && [[ "${TRAVIS_BRANCH}" != master ]]; then
	export REPO="quay.io/coreos/prometheus-operator-dev"
	export REPO_PROMETHEUS_CONFIG_RELOADER="quay.io/coreos/prometheus-config-reloader-dev"
fi

# For both git tags and git branches 'TRAVIS_BRANCH' contains the name.
export TAG="${TRAVIS_BRANCH}"

make image

echo "${QUAY_PASSWORD}" | docker login -u "${QUAY_USERNAME}" --password-stdin quay.io
docker push "${REPO}:${TAG}"
docker push "${REPO_PROMETHEUS_CONFIG_RELOADER}:${TAG}"

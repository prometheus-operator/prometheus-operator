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

CPU_ARCHS="amd64 arm64 arm"

# Push to Quay '-dev' repo if it's not a git tag or master branch build.
export REPO="quay.io/coreos/prometheus-operator"
export REPO_PROMETHEUS_CONFIG_RELOADER="quay.io/coreos/prometheus-config-reloader"

if [[ "${TRAVIS_TAG}" == "" ]] && [[ "${TRAVIS_BRANCH}" != master ]]; then
	export REPO="quay.io/coreos/prometheus-operator-dev"
	export REPO_PROMETHEUS_CONFIG_RELOADER="quay.io/coreos/prometheus-config-reloader-dev"
fi

# For both git tags and git branches 'TRAVIS_BRANCH' contains the name.
export TAG="${TRAVIS_BRANCH}"

for arch in ${CPU_ARCHS}; do
	make --always-make image GOARCH="$arch" TAG="${TAG}-$arch"
done

if [ "$TRAVIS" == "true" ]; then
	# Workaround for docker bug https://github.com/docker/for-linux/issues/396
	sudo chmod o+x /etc/docker
fi

echo "${QUAY_PASSWORD}" | docker login -u "${QUAY_USERNAME}" --password-stdin quay.io
export DOCKER_CLI_EXPERIMENTAL=enabled
for r in ${REPO} ${REPO_PROMETHEUS_CONFIG_RELOADER}; do
	# Images need to be on remote registry before creating manifests
	for arch in $CPU_ARCHS; do
		docker push "${r}:${TAG}-$arch"
	done

	# Create manifest to join all images under one virtual tag
	docker manifest create -a "${r}:${TAG}" \
				  "${r}:${TAG}-amd64" \
				  "${r}:${TAG}-arm64" \
				  "${r}:${TAG}-arm"

	# Annotate to set which image is build for which CPU architecture
	for arch in $CPU_ARCHS; do
		docker manifest annotate --arch "$arch" "${r}:${TAG}" "${r}:${TAG}-$arch"
	done
	docker manifest push "${r}:${TAG}"
done

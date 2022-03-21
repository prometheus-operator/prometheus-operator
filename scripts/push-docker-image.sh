#!/usr/bin/env bash
#
# This script builds multi-arch container images and publish them to multiple container registries.
#
# The script figures out:
# - if an image is a development one (by default suffixed with `-dev` in image name)
# - the image tag (aka version) based on GITHUB_REF value
#
# The script does not:
# - directly execute `docker build`. This is done via the Makefile.
# - Authenticate to the container registries
#

# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail

CPU_ARCHS="amd64 arm64 arm ppc64le s390x"
REGISTRIES="${REGISTRIES:-"quay.io ghcr.io"}"

# IMAGE_OPERATOR, IMAGER_RELOADER and IMAGE_WEBHOOK need to be exported to be used by `make`
export IMAGE_OPERATOR="${IMAGE_OPERATOR:-"prometheus-operator/prometheus-operator"}"
export IMAGE_RELOADER="${IMAGE_RELOADER:-"prometheus-operator/prometheus-config-reloader"}"
export IMAGE_WEBHOOK="${IMAGE_WEBHOOK:="prometheus-operator/admission-webhook"}"
# Figure out if current commit is tagged
export TAG="${GITHUB_REF##*/}"

# Push `-dev` images unless commit is tagged
IMAGE_SUFFIX="-dev"

# Use the main image repository if TAG is a semver tag or it is a main or master branch.
# Otherwise assemble the image tag from VERSION file + short commit SHA and
# push them to the dev image repository.
if [[ "$TAG" =~ ^v[0-9]+\.[0-9]+ ]] || [ "${TAG}" == "master" ] || [ "${TAG}" == "main" ]; then
	# Reset suffixes as images are not development ones
	IMAGE_SUFFIX=""
else
	TAG="v$(cat "$(git rev-parse --show-toplevel)/VERSION")-$(git rev-parse --short HEAD)"
fi

# Compose full image names for retagging and publishing to remote container registries
OPERATORS=""
RELOADERS=""
WEBHOOKS=""
for i in ${REGISTRIES}; do
	OPERATORS="$i/${IMAGE_OPERATOR}${IMAGE_SUFFIX} ${OPERATORS}"
	RELOADERS="$i/${IMAGE_RELOADER}${IMAGE_SUFFIX} ${RELOADERS}"
	WEBHOOKS="$i/${IMAGE_WEBHOOK}${IMAGE_SUFFIX} ${WEBHOOKS}"
done

for img in ${OPERATORS} ${RELOADERS} ${WEBHOOKS}; do
	echo "Building multi-arch image: $img:$TAG"
done

# Build images and rename them for each remote registry
for arch in ${CPU_ARCHS}; do
	make --always-make image GOARCH="$arch" TAG="${TAG}-$arch"
	# Retag operator image
	for i in ${OPERATORS}; do
		docker tag "${IMAGE_OPERATOR}:${TAG}-$arch" "${i}:${TAG}-$arch"
	done
	# Retag reloader image
	for i in ${RELOADERS}; do
		docker tag "${IMAGE_RELOADER}:${TAG}-$arch" "${i}:${TAG}-$arch"
	done
	# Retag webhook image
	for i in ${WEBHOOKS}; do
		docker tag "${IMAGE_WEBHOOK}:${TAG}-$arch" "${i}:${TAG}-$arch"
	done
done

# Compose multi-arch images and push them to remote repositories
export DOCKER_CLI_EXPERIMENTAL=enabled
for r in ${OPERATORS} ${RELOADERS} ${WEBHOOKS}; do
	# Images need to be on remote registry before creating manifests
	for arch in $CPU_ARCHS; do
		docker push "${r}:${TAG}-$arch"
	done

	# Create manifest to join all images under one virtual tag
	docker manifest create -a "${r}:${TAG}" \
				  "${r}:${TAG}-amd64" \
				  "${r}:${TAG}-arm64" \
				  "${r}:${TAG}-arm" \
				  "${r}:${TAG}-ppc64le" \
				  "${r}:${TAG}-s390x"

	# Annotate to set which image is build for which CPU architecture
	for arch in $CPU_ARCHS; do
		docker manifest annotate --arch "$arch" "${r}:${TAG}" "${r}:${TAG}-$arch"
	done
	docker manifest push "${r}:${TAG}"
done

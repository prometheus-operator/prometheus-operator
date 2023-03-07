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

CPU_ARCHS="${CPU_ARCHS:-"amd64 arm64 arm ppc64le s390x"}"
REGISTRIES="${REGISTRIES:-"quay.io ghcr.io"}"

# IMAGE_OPERATOR, IMAGER_RELOADER and IMAGE_WEBHOOK need to be exported to be used by `make`
export IMAGE_OPERATOR="${IMAGE_OPERATOR:-"prometheus-operator/prometheus-operator"}"
export IMAGE_RELOADER="${IMAGE_RELOADER:-"prometheus-operator/prometheus-config-reloader"}"
export IMAGE_WEBHOOK="${IMAGE_WEBHOOK:="prometheus-operator/admission-webhook"}"

# GITHUB_REF and GITHUB_SHA are automatically populated in GitHub actions.
# Otherwise compute them.
COMMIT_SHA="$(echo "${GITHUB_SHA:-$(git rev-parse HEAD)}" | cut -c1-8)"
GITHUB_REF="${GITHUB_REF:-$(git symbolic-ref HEAD)}"
TAG="${GITHUB_REF##*/}"

IMAGE_SUFFIX="-dev"
MAIN_BRANCH=""

# Use the "official" image repository if TAG is a semver tag or it is the main
# branch.
# Otherwise (e.g. release branches), assemble the image tag from VERSION file +
# short commit SHA and push them to the -dev image repository.
if [[ "$TAG" =~ ^v[0-9]+\.[0-9]+ ]] || [ "${TAG}" == "main" ]; then
	# Reset suffixes as images are not development ones
	IMAGE_SUFFIX=""
	if [[ "${TAG}" == "main" ]]; then
		MAIN_BRANCH="yes"
	fi
else
	TAG="v$(cat "$(git rev-parse --show-toplevel)/VERSION")-${COMMIT_SHA}"
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

echo "Tag: ${TAG}"
echo "Main branch: ${MAIN_BRANCH}"
echo "Image suffix: ${IMAGE_SUFFIX}"
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

# Compose the multi-arch images and push them to remote repositories.
export DOCKER_CLI_EXPERIMENTAL=enabled
for r in ${OPERATORS} ${RELOADERS} ${WEBHOOKS}; do
	# Images need to be pushed to the remote registry before creating the manifest.
	MANIFEST="${r}:${TAG}"
	IMAGES=()
	for arch in $CPU_ARCHS; do
		echo "Pushing image ${MANIFEST}-${arch}"
		docker push "${MANIFEST}-${arch}"
		IMAGES[${#IMAGES[@]}]="${MANIFEST}-${arch}"
	done

	# Create the manifest to join all images under one virtual tag.
	echo "Creating manifest ${MANIFEST}"
	docker manifest create --amend "${MANIFEST}" "${IMAGES[@]}"

	# Annotate to set which image is build for which CPU architecture.
	for arch in $CPU_ARCHS; do
		docker manifest annotate --arch "$arch" "${MANIFEST}" "${r}:${TAG}-$arch"
	done

	# Push the manifest to the remote registry.
	echo "Pushing manifest ${MANIFEST}"
	docker manifest push "${MANIFEST}"

	# Sign the manifest for official tags.
	if [[ -z "${MAIN_BRANCH}" ]]; then
		DIGEST="$(crane digest "${MANIFEST}")"
		echo "Signing manifest ${MANIFEST}@${DIGEST}"
		cosign sign --yes -a GIT_HASH="${COMMIT_SHA}" -a GIT_VERSION="${TAG}" "${MANIFEST}@${DIGEST}"
	else
		echo "Not signing the manifest because the tag is 'main'"
	fi
done

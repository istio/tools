#!/usr/bin/env bash

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eux

function wait_for_image() {
  img="${1?image name}"
  SLEEP_TIME=60
  until crane manifest "${img}" > /dev/null; do
    echo "Image ${img} not yet ready, trying again..."
    sleep $SLEEP_TIME
  done
}

# Enable docker buildx
export DOCKER_CLI_EXPERIMENTAL=enabled

# support other container tools, e.g. podman
CONTAINER_CLI=${CONTAINER_CLI:-docker}
# Use buildx for CI by default, allow overriding for old clients or other tools like podman
CONTAINER_BUILDER=${CONTAINER_BUILDER:-"buildx build --load"}
HUB=${HUB:-gcr.io/istio-testing}
# Suffix is derive from the Git SHA we are building for consistency.
# If there is none define, we fallback to date. Note this doesn't work with MANIFEST_ARCH
SUFFIX="${PULL_BASE_SHA:-$(date +%Y-%m-%dT%H-%M-%S)}"
BRANCH=release-1.22
VERSION="${BRANCH}-${SUFFIX}"
SHA="${BRANCH}"
# Arch defines the architecture to tag this image as.
# Note: it is up to the user to ensure that this matches the architecture it was built on; we could add `--platform` explicitly
# but to keep things simple and support alternative builders, we elide it.
# Emulation takes many hours to build, so its recommended to build natively anyways.
ARCH="${TARGET_ARCH:-amd64}"
# MANIFEST_ARCH, if present, defines which architectures we should join together once complete.
# For example, if we have MANIFEST_ARCH="amd64 arm64", after the build is complete we will merge the amd64 and arm64 images
# Generally, this should always be set even with a single architecture build or we will end up with only an image with a `-{arch}` suffix.
MANIFEST_ARCH="${MANIFEST_ARCH-amd64}"

# CACHE_FROM_TAG, if present, defines an image tag which is used in the build --cache-from option.
CACHE_FROM_TAG=":${BRANCH}-latest-${ARCH}"
# The podman build --cache-from option value must contain neither a tag nor digest.
# Overriding of the CACHE_FROM_TAG by setting an empty string when running podman build.
if [[ "${CONTAINER_CLI}" == "podman" ]]; then
  CACHE_FROM_TAG=""
fi

# The docker image runs `go get istio.io/tools@${SHA}`
# In postsubmit, if we pull from the head of the branch, we get a race condition and usually will pull and old version
# In presubmit, this SHA does not exist, so we should just pull from the head of the branch (eg master)
if [[ "${JOB_TYPE:-}" == "postsubmit" ]]; then
  SHA=$(git rev-parse ${BRANCH})
fi

# To generate Docker images on a Mac, setting ADDITIONAL_BUILD_ARGS=--load will result
# in them showing up in `docker images` after running the script.
ADDITIONAL_BUILD_ARGS=${ADDITIONAL_BUILD_ARGS:-}
# Allow overriding of the GOLANG_IMAGE by having it set in the environment
if [[ -n "${GOLANG_IMAGE:-}" ]]; then
  ADDITIONAL_BUILD_ARGS+=" --build-arg GOLANG_IMAGE=${GOLANG_IMAGE}"
fi

# shellcheck disable=SC2086
${CONTAINER_CLI} ${CONTAINER_BUILDER} --target build_tools \
  ${ADDITIONAL_BUILD_ARGS} --build-arg "ISTIO_TOOLS_SHA=${SHA}" --build-arg "VERSION=${VERSION}" \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from "${HUB}/build-tools${CACHE_FROM_TAG}" \
  -t "${HUB}/build-tools:${BRANCH}-latest-${ARCH}" \
  -t "${HUB}/build-tools:${VERSION}-${ARCH}" \
  .

# shellcheck disable=SC2086
${CONTAINER_CLI} ${CONTAINER_BUILDER} --target build_env_proxy \
  ${ADDITIONAL_BUILD_ARGS} --build-arg "ISTIO_TOOLS_SHA=${SHA}" --build-arg "VERSION=${VERSION}" \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from "${HUB}/build-tools-proxy${CACHE_FROM_TAG}" \
  -t "${HUB}/build-tools-proxy:${BRANCH}-latest-${ARCH}" \
  -t "${HUB}/build-tools-proxy:${VERSION}-${ARCH}" \
  .


if [[ -z "${DRY_RUN:-}" ]]; then
  TO_PUSH=(
    "${HUB}/build-tools:${VERSION}"
    "${HUB}/build-tools:${BRANCH}-latest"
    "${HUB}/build-tools-proxy:${VERSION}"
    "${HUB}/build-tools-proxy:${BRANCH}-latest"
  )
  # First, push the architecture specific images
  for image in "${TO_PUSH[@]}"; do
    ${CONTAINER_CLI} push "${image}-${ARCH}"
  done


  # Building the manifest is a bit complex due to limitations in our CI, as typical approachs are not viable:
  # * Emulation is way to slow to build these images (many hours)
  # * Starting up remote buildx builders is plausible, but has portability, security, and complexity concerns
  # * Building each image on a arch-specific job, then "joining" them in a final job would work, but prow cannot do this
  # Instead, we are forced to make one of the jobs the "lead" job, which polls for the other jobs to complete then merges the images.
  if [[ "${MANIFEST_ARCH}" != "" ]]; then
    echo "Build multi-arch manifests for ${MANIFEST_ARCH}"
    IFS=" " read -r -a __arches__ <<< "${MANIFEST_ARCH}"

    for image in "${TO_PUSH[@]}"; do
      images=()
      for arch in "${__arches__[@]}"; do
        arch_img="${image}-${arch}"
        # The other images are pushed by another job, wait for it to be ready
        wait_for_image "${arch_img}"
        images+=("${arch_img}")
      done
      docker manifest rm "${image}" || true
      docker manifest create "${image}" "${images[@]}"
      docker manifest push "${image}"
    done
  fi
fi

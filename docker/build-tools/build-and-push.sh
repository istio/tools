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

# Enable docker buildx
export DOCKER_CLI_EXPERIMENTAL=enabled

# support other container tools, e.g. podman
CONTAINER_CLI=${CONTAINER_CLI:-docker}
# Use buildx for CI by default, allow overriding for old clients or other tools like podman
CONTAINER_BUILDER=${CONTAINER_BUILDER:-"buildx build"}
GOLANG_IMAGE=${GOLANG_IMAGE:-"golang:1.16.4"}
HUB=${HUB:-gcr.io/istio-testing}
DATE=$(date +%Y-%m-%dT%H-%M-%S)
BRANCH=master
VERSION="${BRANCH}-${DATE}"
SHA="${BRANCH}"

BUILD_BUILD_TOOLS=${BUILD_BUILD_TOOLS:-1}
BUILD_BUILD_TOOLS_PROXY=${BUILD_BUILD_TOOLS_PROXY:-1}
BUILD_BUILD_TOOLS_CENTOS=${BUILD_BUILD_TOOLS_CENTOS:-1}

# The docker image runs `go get istio.io/tools@${SHA}`
# In postsubmit, if we pull from the head of the branch, we get a race condition and usually will pull and old version
# In presubmit, this SHA does not exist, so we should just pull from the head of the branch (eg master)
if [[ "${JOB_TYPE:-}" == "postsubmit" ]]; then
  SHA=$(git rev-parse ${BRANCH})
fi


BUILDER_COMMON_FLAGS=${BUILDER_COMMON_FLAGS:-""}
TARGET_PLATFORMS=${TARGET_PLATFORMS:-"linux/amd64"}

if [[ -z "${DRY_RUN:-}" ]]; then
    BUILDER_COMMON_FLAGS="--load"
else
    BUILDER_COMMON_FLAGS="--push"
fi


if [[ ${BUILD_BUILD_TOOLS} == 1 ]]; then
  # shellcheck disable=SC2086
  ${CONTAINER_CLI} ${CONTAINER_BUILDER} ${BUILDER_COMMON_FLAGS} --platform=${TARGET_PLATFORMS} --target build_tools --build-arg "GOLANG_IMAGE=${GOLANG_IMAGE}" --build-arg "ISTIO_TOOLS_SHA=${SHA}" --build-arg "VERSION=${VERSION}" -t "${HUB}/build-tools:${VERSION}" -t "${HUB}/build-tools:${BRANCH}-latest" .
fi

if [[ ${BUILD_BUILD_TOOLS_PROXY} == 1 ]]; then
  # shellcheck disable=SC2086
  ${CONTAINER_CLI} ${CONTAINER_BUILDER} ${BUILDER_COMMON_FLAGS} --platform=${TARGET_PLATFORMS} --build-arg "GOLANG_IMAGE=${GOLANG_IMAGE}" --build-arg "ISTIO_TOOLS_SHA=${SHA}" --build-arg "VERSION=${VERSION}" -t "${HUB}/build-tools-proxy:${VERSION}" -t "${HUB}/build-tools-proxy:${BRANCH}-latest" .
fi

if [[ ${BUILD_BUILD_TOOLS_CENTOS} == 1 ]]; then
  # shellcheck disable=SC2086
  ${CONTAINER_CLI} ${CONTAINER_BUILDER} ${BUILDER_COMMON_FLAGS} --platform=${TARGET_PLATFORMS} --build-arg "ISTIO_TOOLS_SHA=${SHA}" --build-arg "VERSION=${VERSION}" -t "${HUB}/build-tools-centos:${VERSION}" -t "${HUB}/build-tools-centos:${BRANCH}-latest" -f Dockerfile.centos .
fi

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

# support other container tools, e.g. podman
CONTAINER_CLI=${CONTAINER_CLI:-docker}

HUB=${HUB:-gcr.io/istio-testing}
DATE=$(date +%Y-%m-%dT%H-%M-%S)
BRANCH=master
VERSION="${BRANCH}-${DATE}"
SHA=$(git rev-parse ${BRANCH})

${CONTAINER_CLI} build --target build_tools --build-arg "ISTIO_TOOLS_SHA=${SHA}" -t "${HUB}/build-tools:${VERSION}" -t "${HUB}/build-tools:${BRANCH}-latest" .
${CONTAINER_CLI} build --build-arg "ISTIO_TOOLS_SHA=${SHA}" -t "${HUB}/build-tools-proxy:${VERSION}" -t "${HUB}/build-tools-proxy:${BRANCH}-latest" .

if [[ -z "${DRY_RUN}" ]]; then
  ${CONTAINER_CLI} push "${HUB}/build-tools:${VERSION}"
  ${CONTAINER_CLI} push "${HUB}/build-tools:${BRANCH}-latest"
  ${CONTAINER_CLI} push "${HUB}/build-tools-proxy:${VERSION}"
  ${CONTAINER_CLI} push "${HUB}/build-tools-proxy:${BRANCH}-latest"
fi

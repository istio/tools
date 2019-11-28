#!/bin/bash

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

set -ex

# shellcheck disable=SC2086
WD=$(cd $(dirname $0); pwd)

function setup_test() {
  local DIRNAME="${1:?"test directory"}"
  local NAMESPACE="istio-stability-${NAMESPACE:-"$1"}"
  local HELM_ARGS="${2:-}"

  mkdir -p "${WD}/tmp"
  local OUTFILE="./perf/stability/tmp/${DIRNAME}.yaml"

  # shellcheck disable=SC2086
  helm --namespace "${NAMESPACE}" template "${WD}/${DIRNAME}" ${HELM_ARGS} > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
    kubectl create ns "${NAMESPACE}" || true
    kubectl label namespace "${NAMESPACE}" istio-injection=enabled || true

    kubectl -n "${NAMESPACE}" apply -f "${OUTFILE}"
  fi
}

setup_test "$@"

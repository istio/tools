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


#set -ex

WD=$(dirname "$0")
WD=$(cd "${WD}" || exit; pwd)

function setup_test() {
  local NAMESPACE="${NAMESPACE:-"pilot-load"}"
  local HELM_FLAGS=${HELM_FLAGS:-"instances=50"}

  mkdir -p "${WD}/tmp"
  local OUTFILE="${WD}/tmp/${NAMESPACE}.yaml"

  kubectl create ns "${NAMESPACE}" || true
  kubectl label namespace "${NAMESPACE}" istio-injection=enabled || true

  helm --namespace "${NAMESPACE}" --set "${HELM_FLAGS}" template "${WD}" > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl --namespace "${NAMESPACE}" apply -f "${OUTFILE}"
  fi
}

setup_test

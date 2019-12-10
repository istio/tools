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
  local NAMESPACE="mtls"
  local HELM_FLAGS=${HELM_FLAGS:-""}

  mkdir -p "${WD}/tmp"
  local OUTFILE="${WD}/tmp/${NAMESPACE}.yaml"

  kubectl create ns "${NAMESPACE}" || true
  kubectl label namespace "${NAMESPACE}" istio-injection=enabled || true

  helm --namespace "${NAMESPACE}" -f ./values.yaml --set "${HELM_FLAGS}" \
    template "${WD}" > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl --namespace "${NAMESPACE}" apply -f "${OUTFILE}"
      pushd ../../load/loadclient || exit
      # Generate larger load for load client.
      export LOADCLIENT_EXTRA_HELM_FLAGS="--set qps=200"
      ./setup_test.sh "${NAMESPACE}" "svc-"
      popd || exit
  fi
}

function setup_istio() {
  pushd ../../istio-install || exit
  export OPERATOR_PROFILE="automtls.yaml" && ./setup_istio_operator.sh
  popd || exit
}

# setup_istio
# setup_test

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

# Usage
# ./setup_istio_operator.sh, install Istio with operator.
#
# export OPERATOR_SHA="abcdef" && ./setup_istio_operator.sh.
# setup Istio with a given operator repo's SHA
#
# export OPERATOR_PROFILE="automtls.yaml" && ./setup_istio_operator.sh.
# using a specific Istio operator profile to install Istio.

set -ex

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
DIRNAME="${WD}/tmp"
mkdir -p "${DIRNAME}"
export GO111MODULE=on

# The profile containing IstioControlPlane spec. Overriding this environment
# variable allow to specify different installation options.
OPERATOR_SHA=${OPERATOR_SHA-$(cat "${WD}"/istio_operator.sha)}
OPERATOR_PROFILE=${OPERATOR_PROFILE:-operator_default.yaml}

ISTIO_OPERATOR_DIR="${DIRNAME}/operator"
if [[ ! -d "${ISTIO_OPERATOR_DIR}" ]]; then
  git clone https://github.com/istio/operator.git "$ISTIO_OPERATOR_DIR"
fi

pushd .
cd "${ISTIO_OPERATOR_DIR}"
git fetch
echo "Checking out operator SHA ${OPERATOR_SHA} ..."
git checkout "${OPERATOR_SHA}"
popd

defaultNamespace=istio-system


function setup_admin_binding() {
  kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user="$(gcloud config get-value core/account)" || true
}

function install_istio() {
    local CR_FILENAME="${WD}/${OPERATOR_PROFILE}"
    pushd "${ISTIO_OPERATOR_DIR}"
    go run ./cmd/mesh.go manifest apply -f "${CR_FILENAME}" --force=true --set defaultNamespace=${defaultNamespace}
    popd
    echo "installation is done"
}

setup_admin_binding
install_istio

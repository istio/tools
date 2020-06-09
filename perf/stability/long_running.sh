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

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")

set -eux

# download latest istio release and install
export TAG="latest"
export DNS_DOMAIN="fake-dns.org"
${ROOT}/istio-install/setup_istio.sh

# setup service graph
pushd "${ROOT}/load"
# shellcheck disable=SC1090
source "./common.sh"
NAMESPACE_NUM=2
START_NUM=0
export DELETE=""
export CMD=""
export WD="${ROOT}/load"
start_servicegraphs "${NAMESPACE_NUM}" "${START_NUM}"
popd

export NOT_INJECTED="True"
# deploy alertmanager related resources
go run ./alertmanager/webhook.go &>/dev/null &
./setup_test.sh alertmanager

# deploy canary upgrader
kubectl create configmap canary-script --from-file=./canary-upgrader/canary_upgrade.sh
./setup_test.sh canary-upgrader


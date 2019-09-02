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
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -ex

NAMESPACE=${1:?"namespace"}
NAMEPREFIX=${2:?"prefix name for service. typically svc-"}

HTTPS=${HTTPS:-"false"}

if [[ -z "${GATEWAY_URL}" ]];then
SYSTEM_GATEWAY_URL=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}' || true)
INGRESS_GATEWAY_URL=$(kubectl -n istio-ingress get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}' || true)
GATEWAY_URL=${SYSTEM_GATEWAY_URL:-$INGRESS_GATEWAY_URL}
fi

SERVICEHOST="${NAMEPREFIX}0.local"

function run_test() {
  YAML=$(mktemp).yml
  helm -n "${NAMESPACE}" template \
	  --set serviceHost="${SERVICEHOST}" \
    --set Namespace="${NAMESPACE}" \
    --set ingress="${GATEWAY_URL}" \
    --set domain="${DNS_DOMAIN}" \
    --set https="${HTTPS}" \
          . > "${YAML}"
  echo "Wrote ${YAML}"

  if [[ -z "${DELETE}" ]];then
    kubectl create ns "${NAMESPACE}" || true
    kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite
    kubectl label namespace "${NAMESPACE}" istio-env=istio-control --overwrite
    sleep 5
    kubectl -n "${NAMESPACE}" apply -f "${YAML}"
  else
    kubectl -n "${NAMESPACE}" delete -f "${YAML}"
  fi
}

run_test

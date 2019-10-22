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

if [[ -z "${GATEWAY_URL:-}" ]];then
SYSTEM_GATEWAY_URL=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}' || true)
INGRESS_GATEWAY_URL=$(kubectl -n istio-ingress get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}' || true)
GATEWAY_URL=${SYSTEM_GATEWAY_URL:-$INGRESS_GATEWAY_URL}
fi

HTTPS=${HTTPS:-"false"}
H2UPGRADE=${H2UPGRADE:-"false"}

function run_test() {
  local ns=${1:?"namespaces"}
  local prefix=${2:?"prefix name for service. typically svc-"}

   YAML=$(mktemp).yml
  # shellcheck disable=SC2086
  helm -n ${ns} template \
          --set serviceNamePrefix="${prefix}" \
          --set Namespace="${ns}" \
          --set domain="${DNS_DOMAIN}" \
          --set ingress="${GATEWAY_URL}" \
          --set https="${HTTPS}" \
          --set h2upgrade="${H2UPGRADE}" \
          . > "${YAML}"
  echo "Wrote ${YAML}"

  kubectl create ns "${ns}" || true
  kubectl label namespace "${ns}" istio-injection=enabled --overwrite
  kubectl label namespace "${ns}" istio-env=istio-control --overwrite

   if [[ -z "${DELETE}" ]];then
    sleep 3
    kubectl -n "${ns}" apply -f "${YAML}"
  else
    kubectl -n "${ns}" delete -f "${YAML}" || true
    kubectl delete ns "${ns}"
  fi
}

function start_servicegraphs() {
  local nn=${1:?"number of namespaces"}
  local min=${2:-"0"}

   # shellcheck disable=SC2004
   for ((ii=$min; ii<$nn; ii++)) {
    ns=$(printf 'service-graph%.2d' $ii)
    prefix=$(printf 'svc%.2d-' $ii)
    if [[ -z "${DELETE}" ]];then
      ${CMD} run_test "${ns}" "${prefix}"
      ${CMD} "${WD}/loadclient/setup_test.sh" "${ns}" "${prefix}"
    else
      ${CMD} "${WD}/loadclient/setup_test.sh" "${ns}" "${prefix}"
      ${CMD} run_test "${ns}" "${prefix}"
    fi

    sleep 30
  }
}

 # Get pod ip range, there must be a better way, but this works.
function ip_range() {
    kubectl get pods --namespace kube-system -o wide | grep kube-dns | awk '{print $6}'|head -1 | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'
}

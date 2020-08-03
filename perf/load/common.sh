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
  GATEWAY_URL=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}' || true)
fi

HTTPS=${HTTPS:-"false"}
H2UPGRADE=${H2UPGRADE:-"false"}

function run_test() {
  local ns=${1:?"namespaces"}
  local prefix=${2:?"prefix name for service. typically svc-"}
  MULTI_CLUSTER=${MULTI_CLUSTER:-"false"}
  CLUSTER1=${CLUSTER1:-"false"}
  CLUSTER2=${CLUSTER2:-"false"}
  YAML=$(mktemp).yml
  # shellcheck disable=SC2086
  helm -n ${ns} template \
          --set serviceNamePrefix="${prefix}" \
          --set Namespace="${ns}" \
          --set domain="${DNS_DOMAIN}" \
          --set ingress="${GATEWAY_URL}" \
          --set https="${HTTPS}" \
          --set h2upgrade="${H2UPGRADE}" \
          --set multicluster.enabled="${MULTI_CLUSTER}" \
          --set multicluster.cluster1="${CLUSTER1}" \
          --set multicluster.cluster2="${CLUSTER2}" \
          . > "${YAML}"
  echo "Wrote ${YAML}"

  kubectl create ns "${ns}" || true
  kubectl label namespace "${ns}" "${INJECTION_LABEL:-istio-injection=enabled}" --overwrite

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
    ns=$(printf 'service-graph%.2d' "${ii}")
    prefix=$(printf 'svc%.2d-' "${ii}")
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

# This function assumes preconfigured ASM multi clusters ready.
# Here is a simple setup guide for example: https://cloud.google.com/service-mesh/docs/gke-install-multi-cluster
function start_servicegraphs_multicluster() {
  local nn=${1:?"number of namespaces"}
  local min=${2:-"0"}
  if [[ -z "${CTX1}" ]] || [[ -z "${CTX2}" ]];then
     echo "Required to set CTX1 and CTX2 env variables first for two clusters"
     exit 1
  fi
   # shellcheck disable=SC2004
   for ((ii=$min; ii<$nn; ii++)) {
    ns=$(printf 'service-graph%.2d' $ii)
    prefix=$(printf 'svc%.2d-' $ii)
    if [[ -z "${DELETE}" ]];then
      kubectl config use-context "${CTX1}"
      MULTI_CLUSTER="true" CLUSTER1="true" CLUSTER2="false" ${CMD} run_test "${ns}" "${prefix}"
      ${CMD} "${WD}/loadclient/setup_test.sh" "${ns}" "${prefix}"

      kubectl config use-context "${CTX2}"
      MULTI_CLUSTER="true" CLUSTER1="false" CLUSTER2="true" ${CMD} run_test "${ns}" "${prefix}"
      ${CMD} "${WD}/loadclient/setup_test.sh" "${ns}" "${prefix}"
      kubectl config use-context "${CTX1}"
    else
      ${CMD} "${WD}/loadclient/setup_test.sh" "${ns}" "${prefix}"
      ${CMD} run_test "${ns}" "${prefix}"
    fi

    sleep 30
  }
}

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

# shellcheck disable=SC2086
WD=$(dirname $0)
# shellcheck disable=SC2164
WD=$(cd "${WD}"; pwd)
# shellcheck disable=SC2164
cd "${WD}"

set -x
NAMESPACE=${NAMESPACE:-'twopods'}
DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org or local"}
TMPDIR=${TMPDIR:-${WD}/tmp}
RBAC_ENABLED="false"
LINKERD_INJECT="${LINKERD_INJECT:-'disabled'}"
echo "linkerd inject is ${LINKERD_INJECT}"

mkdir -p "${TMPDIR}"

# Get pod ip range, there must be a better way, but this works.
function pod_ip_range() {
    kubectl get pods --namespace kube-system -o wide | grep kube-dns | awk '{print $6}'|head -1 | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'
}

function svc_ip_range() {
    kubectl -n kube-system get svc kube-dns --no-headers | awk '{print $3}' | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'
}

function run_test() {
  # shellcheck disable=SC2046
  helm -n ${NAMESPACE} template \
      --set rbac.enabled="${RBAC_ENABLED}" \
      --set includeOutboundIPRanges=$(svc_ip_range) \
      --set injectL="${LINKERD_INJECT}" \
      --set domain="${DNS_DOMAIN}" \
          . > ${TMPDIR}/twopods.yaml
  echo "Wrote ${TMPDIR}/twopods.yaml"

  # remove stdio rules
  kubectl apply -n ${NAMESPACE} -f ${TMPDIR}/twopods.yaml
  echo ${TMPDIR}/twopods.yaml
}

for ((i=1; i<=$#; i++)); do
    case ${!i} in
        -r|--rbac) ((i++)); RBAC_ENABLED="true"
        continue
        ;;
    esac
done
kubectl create ns ${NAMESPACE} || true
kubectl label namespace ${NAMESPACE} istio-injection=enabled --overwrite || true
if [[ "$LINKERD_INJECT" == "enabled" ]]
then
  kubectl annotate namespace ${NAMESPACE} linkerd.io/inject=enabled || true
fi
run_test

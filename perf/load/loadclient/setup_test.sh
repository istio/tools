#!/bin/bash
WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -ex

NAMESPACE=${1:?"namespace"}
NAMEPREFIX=${2:?"prefix name for service. typically svc-"}

HTTPS=${HTTPS:-"false"}

SYSTEM_GATEWAY_URL=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}' || true)
INGRESS_GATEWAY_URL=$(kubectl -n istio-ingress get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}' || true)
GATEWAY_URL=${SYSTEM_GATEWAY_URL:-$INGRESS_GATEWAY_URL}
SERVICEHOST="${NAMEPREFIX}0.local"

function run_test() {
  YAML=$(mktemp).yml
  helm -n ${NAMESPACE} template \
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
    sleep 5
    kubectl -n "${NAMESPACE}" apply -f "${YAML}"
  else
    kubectl -n "${NAMESPACE}" delete -f "${YAML}"
  fi
}

run_test

#!/bin/bash
WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -ex

NAMESPACE=${1:?"namespace"}
NAMEPREFIX=${2:?"prefix name for service. typically svc-"}

GATEWAY_URL=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
SERVICEHOST="${NAMEPREFIX}0.local"

function run_test() {
  YAML=$(mktemp).yml
  helm -n ${NAMESPACE} template \
	  --set serviceHost="${SERVICEHOST}" \
    --set Namespace="${NAMESPACE}" \
    --set ingress="${GATEWAY_URL}" \
          . > "${YAML}"
  echo "Wrote ${YAML}"

  # remove stdio rules
  kubectl --namespace istio-system delete rules stdio stdiotcp || true
  
  
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

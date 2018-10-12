#!/bin/bash
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

  kubectl create ns "${NAMESPACE}" || true
  kubectl label namespace "${NAMESPACE}" istio-injection=enabled --overwrite

  # remove stdio rules
  kubectl --namespace istio-system delete rules stdio stdiotcp || true

  sleep 5
  kubectl -n "${NAMESPACE}" apply -f "${YAML}"
}

run_test

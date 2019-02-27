#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

function setup_test() {
  local DIRNAME="${1:?"test directory"}"
  local NAMESPACE="${NAMESPACE:-"$1"}"
  local HELM_ARGS="${2:-}"

  mkdir -p "${WD}/tmp"
  local OUTFILE="${WD}/tmp/${DIRNAME}.yaml"

  kubectl create ns "${NAMESPACE}" || true
  kubectl label namespace "${NAMESPACE}" istio-injection=enabled || true

  helm -n "${NAMESPACE}" template "${DIRNAME}" ${HELM_ARGS} > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl -n "${NAMESPACE}" apply -f "${OUTFILE}"
  fi
}

function delete_test() {
  local DIRNAME="${1:?"test directory"}"
  local NAMESPACE="${NAMESPACE:-"$1"}"
  local OUTFILE="${WD}/tmp/${DIRNAME}.yaml"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl -n "${NAMESPACE}" delete -f "${OUTFILE}"
  fi
}


function delete_tests() {
  for test in $1; do
      delete_test "${test}"
  done
}

function setup_tests() {
  for test in $1; do
    case "${test}" in
      "allconfig") setup_all_config;;
      "gateway-bouncer") setup_gateway_bouncer;;
      *) setup_test "${test}";;
    esac
  done
}

function setup_all_config() {
  local gateway=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  local domain=${DNS_DOMAIN:-qualistio.org}
  setup_test "allconfig" "--set ingress=${gateway} --set domain=${domain}"
}

function setup_gateway_bouncer() {
  setup_test "gateway-bouncer" "--set namespace=${NAMESPACE:-"$test"}"
  local NAMESPACE="${NAMESPACE:-"$test"}"
  # Waiting until LoadBalancer is created and retrieving the assigned
  # external IP address.
  while : ; do
    INGRESS_IP=$(kubectl -n ${NAMESPACE} \
      get service istio-ingress-${NAMESPACE} \
      -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

    if [[ -z "${INGRESS_IP}" ]]; then
      sleep 5s
    else
      break
    fi
  done

  # Populating a ConfigMap with the external IP address and restarting the
  # client to pick up the new version of the ConfigMap.
  kubectl -n ${NAMESPACE} delete configmap fortio-client-config
  kubectl -n ${NAMESPACE} create configmap fortio-client-config \
    --from-literal=external_addr=${INGRESS_IP}
  kubectl -n ${NAMESPACE} patch deployment fortio-client \
    -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"date\":\"`date +'%s'`\"}}}}}"
}

ALL_TESTS="http10 graceful-shutdown gateway-bouncer"
TESTS="${TESTS:-"$ALL_TESTS"}"

case "$1" in
   "") echo "Pass one of setup or delete" ;;
  "setup" | "install") setup_tests "${TESTS}" ;;
  "delete" | "remove" | "uninstall") delete_tests "${TESTS}" ;;
esac
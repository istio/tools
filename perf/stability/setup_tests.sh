#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

source "${WD}/common_setup.sh"

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
      delete_test "${test}" || true
  done
}

function setup_tests() {
  for test in $1; do
    case "${test}" in
      "allconfig") setup_all_config;;
      "gateway-bouncer") ${WD}/gateway-bouncer/setup.sh;;
      "sds-certmanager") ${WD}/sds-certmanager/setup.sh;;
      *) setup_test "${test}";;
    esac
  done
}

function setup_all_config() {
  local gateway=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
  local domain=${DNS_DOMAIN:-qualistio.org}
  setup_test "allconfig" "--set ingress=${gateway} --set domain=${domain}"
}

ALL_TESTS="http10 graceful-shutdown gateway-bouncer mysql"
TESTS="${TESTS:-"$ALL_TESTS"}"

case "$1" in
   "") echo "Pass one of setup or delete" ;;
  "setup" | "install") setup_tests "${TESTS}" ;;
  "delete" | "remove" | "uninstall") delete_tests "${TESTS}" ;;
esac
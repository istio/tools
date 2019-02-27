#!/bin/bash

WD=$(dirname $0)
WD=$(cd $WD; pwd)

function setup_test() {
  local DIRNAME="${1:?"test directory"}"
  local NAMESPACE="${NAMESPACE:-"$1"}"

  mkdir -p "${WD}/tmp"
  local OUTFILE="${WD}/tmp/${DIRNAME}.yaml"

  kubectl create ns "${NAMESPACE}" || true
  kubectl label namespace "${NAMESPACE}" istio-injection=enabled || true

  helm -n "${NAMESPACE}" template "${DIRNAME}" > "${OUTFILE}"

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
      setup_test "${test}"
  done
}

ALL_TESTS="http10 graceful-shutdown"
TESTS="${TESTS:-"$ALL_TESTS"}"

case "$1" in
   "")
   echo "Pass one of setup or delete"
   ;;
  "setup" | "install")
    setup_tests "${TESTS}"
    ;;
  "delete" | "remove" | "uninstall")
    delete_tests "${TESTS}"
    ;;
esac
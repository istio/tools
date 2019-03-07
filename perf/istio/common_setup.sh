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

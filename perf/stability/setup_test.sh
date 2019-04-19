#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

function setup_test() {
  local DIRNAME="${1:?"test directory"}"
  local NAMESPACE="istio-stability-${NAMESPACE:-"$1"}"
  local HELM_ARGS="${2:-}"

  mkdir -p "${WD}/tmp"
  local OUTFILE="${WD}/tmp/${DIRNAME}.yaml"

  helm --namespace "${NAMESPACE}" template "${WD}/${DIRNAME}" ${HELM_ARGS} > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
    kubectl create ns "${NAMESPACE}" || true
    kubectl label namespace "${NAMESPACE}" istio-injection=enabled || true

    kubectl -n "${NAMESPACE}" apply -f "${OUTFILE}"
  fi
}

setup_test "$@"
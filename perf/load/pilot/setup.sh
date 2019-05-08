#!/bin/bash

#set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

function setup_test() {
  local NAMESPACE="${NAMESPACE:-"pilot-load"}"

  mkdir -p "${WD}/tmp"
  local OUTFILE="${WD}/tmp/${NAMESPACE}.yaml"

  kubectl create ns "${NAMESPACE}" || true
  kubectl label namespace "${NAMESPACE}" istio-injection=enabled || true

  helm --namespace "${NAMESPACE}" "${HELM_FLAGS}" template "${WD}" > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl --namespace "${NAMESPACE}" apply -f "${OUTFILE}"
  fi
}

setup_test

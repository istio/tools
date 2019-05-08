#!/bin/bash

#set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

function setup_test() {
  local NAMESPACE="${NAMESPACE:-"pilot-load"}"
  local HELM_FLAGS=${HELM_FLAGS:-"instances=50"}

  mkdir -p "${WD}/tmp"
  local OUTFILE="${WD}/tmp/${NAMESPACE}.yaml"

  kubectl create ns "${NAMESPACE}" || true
  kubectl label namespace "${NAMESPACE}" istio-injection=enabled || true

  helm --namespace "${NAMESPACE}" --set "${HELM_FLAGS}" template "${WD}" > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl --namespace "${NAMESPACE}" apply -f "${OUTFILE}"
  fi
}

setup_test

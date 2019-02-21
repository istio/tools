#!/bin/bash

set -ex

NAMESPACE="${NAMESPACE:-istio-upgrader}"

function install_all_config() {
  local DIRNAME="${1:?"output dir"}"
  local OUTFILE="${DIRNAME}/all_config.yaml"

  kubectl create ns $NAMESPACE || true

  helm --set "namespace=$NAMESPACE" -n $NAMESPACE template . > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl -n $NAMESPACE apply -f "${OUTFILE}"
  fi
}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

install_all_config "${WD}/tmp" $*

#!/bin/bash

set -ex

function install_all_config() {
  local DIRNAME="${1:?"output dir"}"
  local OUTFILE="${DIRNAME}/all_config.yaml"

  kubectl create ns graceful-shutdown || true

  kubectl label namespace graceful-shutdown istio-injection=enabled || true

  helm -n graceful-shutdown template . > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl -n graceful-shutdown apply -f "${OUTFILE}"
  fi
}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

install_all_config "${WD}/tmp" $*

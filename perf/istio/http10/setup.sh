#!/bin/bash

set -ex

function install_all_config() {
  local DIRNAME="${1:?"output dir"}"
  local OUTFILE="${DIRNAME}/all_config.yaml"

  kubectl create ns http10 || true

  kubectl label namespace http10 istio-injection=enabled || true

  helm -n http10 template . > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl -n http10 apply -f "${OUTFILE}"
  fi
}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

install_all_config "${WD}/tmp" $*

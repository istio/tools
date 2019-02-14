#!/bin/bash

set -ex

function install_all_config() {
  local DIRNAME="${1:?"output dir"}"
  local OUTFILE="${DIRNAME}/all_config.yaml"

  kubectl create ns allow-external-traffic-a || true
  kubectl label namespace allow-external-traffic-a istio-injection=enabled || true

  kubectl create ns allow-external-traffic-b || true
  kubectl label namespace allow-external-traffic-b istio-injection=enabled || true

  INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

  helm template . --set "externalDestination=http://${INGRESS_HOST:?"Ingress could not be found"}" > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl apply -f "${OUTFILE}"
  fi
}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

install_all_config "${WD}/tmp" $*

# Must set global.outboundTrafficPolicy.mode=ALLOW_ANY

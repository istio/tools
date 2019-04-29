#!/bin/bash
set -ex

DNS_DOMAIN=${DNS_DOMAIN:-local}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

HUB="${1:?"build hub"}"
TAG="${2:?"build tag"}"
GOPATH="${GOPATH:?go path is required}"
INSTALLER="${GOPATH}/src/github.com/istio-ecosystem/istio-installer"


function setup_admin_binding() {
  kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account) || true
}

function iop() {
  export HUB=$HUB
  BASE="${INSTALLER}" HUB=${HUB} TAG=${TAG} ${INSTALLER}/bin/iop $* --values "${WD}/values-new-installer.yaml"
}

function install_istio() {
  local DIRNAME="${1:?"output dir"}"
  local opts=""

  if [[ -z "${SMALL_INSTALL}" ]]; then
    opts+="--values values-large.yaml"
  fi

  mkdir -p ${DIRNAME}/
  mkdir -p ${DIRNAME}/control
  mkdir -p ${DIRNAME}/telemetry
	cp -aR ${INSTALLER}/crds/ ${DIRNAME}/crds
	iop istio-system istio-system-security ${INSTALLER}/security/citadel -t ${opts} > ${DIRNAME}/citadel.yaml
	iop istio-control istio-config ${INSTALLER}/istio-control/istio-config -t ${opts} > ${DIRNAME}/control/istio-config.yaml
	iop istio-control istio-discovery ${INSTALLER}/istio-control/istio-discovery -t ${opts} > ${DIRNAME}/control/istio-discovery.yaml
	iop istio-control istio-autoinject ${INSTALLER}/istio-control/istio-autoinject -t ${opts} > ${DIRNAME}/control/istio-autoinject.yaml
	iop istio-ingress istio-ingress ${INSTALLER}/gateways/istio-ingress -t ${opts} > ${DIRNAME}/istio-ingress.yaml
	iop istio-egress istio-egress ${INSTALLER}/gateways/istio-egress -t ${opts} > ${DIRNAME}/istio-egress.yaml
	iop istio-telemetry istio-telemetry ${INSTALLER}/istio-telemetry/mixer-telemetry -t ${opts} > ${DIRNAME}/telemetry/istio-telemetry.yaml
	iop istio-telemetry istio-grafana ${INSTALLER}/istio-telemetry/grafana -t ${opts} > ${DIRNAME}/telemetry/istio-grafana.yaml
	iop istio-policy istio-policy ${INSTALLER}/istio-policy -t ${opts} > ${DIRNAME}/istio-policy.yaml


  if [[ -z "${DRY_RUN}" ]]; then
    kubectl create namespace istio-system || true
    kubectl create namespace istio-control || true
    kubectl create namespace istio-ingress || true
    kubectl create namespace istio-telemetry || true

    kubectl apply -f "${DIRNAME}/crds/"
    kubectl wait --for=condition=Established -f "${DIRNAME}/crds/"

    kubectl apply -f "${DIRNAME}/citadel.yaml"
    kubectl rollout status deployment istio-citadel11 -n istio-system --timeout=1m

    kubectl apply -f "${DIRNAME}/control/"
    kubectl rollout status deployment istio-galley -n istio-control --timeout=1m
    kubectl rollout status deployment istio-pilot  -n istio-control --timeout=1m
    kubectl rollout status deployment istio-sidecar-injector -n istio-control --timeout=1m

    kubectl apply -f "${DIRNAME}/istio-ingress.yaml"
    kubectl rollout status deployment ingressgateway -n istio-ingress --timeout=1m

    kubectl apply -f "${DIRNAME}/telemetry/"
    kubectl rollout status deployment istio-telemetry -n istio-telemetry --timeout=1m
    kubectl rollout status deployment grafana -n istio-telemetry --timeout=1m
  fi
}

function install_gateways() {
  local domain=${DNS_DOMAIN:-qualistio.org}
  if [[ -z "${DRY_RUN}" ]]; then
      helm template --set domain="${domain}" "${WD}/base" | kubectl -n istio-system apply -f -
  fi
}

setup_admin_binding
install_istio "${WD}/tmp"
$WD/setup_prometheus.sh "${WD}/tmp"
install_gateways

#!/bin/bash
set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

function install_prometheus() {
  local DIRNAME="$1" # should be like tools/perf/istio/tmp
  DIRNAME=$(cd $DIRNAME; pwd)
  INSTALLER_DIR="${DIRNAME}/installer"

  git clone https://github.com/istio/installer.git  "$INSTALLER_DIR"

  # Create GCP SSD Storage class for Prometheus to use. May not work outside GKE
  kubectl apply -f "${WD}/../prometheus-install/ssd-storage-class.yaml"

  kubectl create ns istio-prometheus || true
  kubectl label ns istio-prometheus istio-injection=disabled --overwrite
  curl -s https://raw.githubusercontent.com/coreos/prometheus-operator/v0.31.1/bundle.yaml | sed "s/namespace: default/namespace: istio-prometheus/g" | kubectl apply -f -
  kubectl -n istio-prometheus wait --for=condition=available --timeout=60s deploy/prometheus-operator
  # kubectl wait is problematic, as the CRDs may not exist before the command is issued.
  until timeout 60s kubectl get crds/prometheuses.monitoring.coreos.com; do echo "Waiting for CRDs to be created..."; done
  until timeout 60s kubectl get crds/alertmanagers.monitoring.coreos.com; do echo "Waiting for CRDs to be created..."; done
  until timeout 60s kubectl get crds/podmonitors.monitoring.coreos.com; do echo "Waiting for CRDs to be created..."; done
  until timeout 60s kubectl get crds/prometheusrules.monitoring.coreos.com; do echo "Waiting for CRDs to be created..."; done
  until timeout 60s kubectl get crds/servicemonitors.monitoring.coreos.com; do echo "Waiting for CRDs to be created..."; done
  helm template --namespace istio-prometheus "${INSTALLER_DIR}"/istio-telemetry/prometheus-operator/ -f "${INSTALLER_DIR}"/global.yaml --set prometheus.createPrometheusResource=true | kubectl apply -n istio-prometheus -f -

  kubectl delete pod -l app=grafana
}

install_prometheus $1

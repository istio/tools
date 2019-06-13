#!/bin/bash
set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

function install_prometheus() {
  local DIRNAME="$1" # should be like tools/perf/istio/tmp
  DIRNAME=$(cd $DIRNAME; pwd)
  local PROMETHEUS_INSTALL="${DIRNAME}"

  # Create GCP SSD Storage class for Prometheus to use. May not work outside GKE
  kubectl apply -f "${WD}/../prometheus-install/ssd-storage-class.yaml"

  helm fetch stable/prometheus-operator --untar --untardir "${PROMETHEUS_INSTALL}"

  kubectl create ns istio-prometheus || true
  PROM_OP="${DIRNAME}prom-operator.yaml"
  helm template "${PROMETHEUS_INSTALL}/prometheus-operator/" \
    --namespace istio-prometheus \
    --name prometheus \
    -f "${WD}/../prometheus-install/values.yaml" > "${PROM_OP}"

  echo "${PROM_OP}"

  if [[ -z "${DRY_RUN}" ]]; then
    kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/alertmanager.crd.yaml
    kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/prometheus.crd.yaml
    kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/prometheusrule.crd.yaml
    kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/servicemonitor.crd.yaml

    kubectl wait --for=condition=Established -f https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/alertmanager.crd.yaml
    kubectl wait --for=condition=Established -f https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/prometheus.crd.yaml
    kubectl wait --for=condition=Established -f https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/prometheusrule.crd.yaml
    kubectl wait --for=condition=Established -f https://raw.githubusercontent.com/coreos/prometheus-operator/master/example/prometheus-operator-crd/servicemonitor.crd.yaml

    kubectl apply --namespace istio-prometheus -f "${PROM_OP}"

    # delete grafana pod so it redeploys with new config
    kubectl delete pod -l app=grafana
  fi
}

install_prometheus $1

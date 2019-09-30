#!/bin/bash

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -ex

# shellcheck disable=SC2086
WD=$(dirname $0)
# shellcheck disable=SC2086
WD=$(cd $WD; pwd)
# shellcheck disable=SC2086
mkdir -p "${WD}/tmp"

function install_prometheus() {
  local DIRNAME="$1" # should be like tools/perf/istio-install/tmp
  # shellcheck disable=SC2086
  DIRNAME=$(cd $DIRNAME; pwd)
  INSTALLER_DIR="${DIRNAME}/installer"
  if [[ ! -e "$INSTALLER_DIR" ]]; then
    git clone https://github.com/istio/installer.git "$INSTALLER_DIR"
  fi

  # Create GCP SSD Storage class for Prometheus to use. May not work outside GKE
  kubectl apply -f "${WD}/../prometheus-install/ssd-storage-class.yaml"

  kubectl create ns istio-prometheus || true
  kubectl label ns istio-prometheus istio-injection=disabled --overwrite
  curl -s https://raw.githubusercontent.com/coreos/prometheus-operator/v0.31.1/bundle.yaml | sed "s/namespace: default/namespace: istio-prometheus/g" | kubectl apply -f -

  kubectl -n istio-prometheus wait --for=condition=available --timeout=60s deploy/prometheus-operator

  # Check CRD
  CMDs_ARR=('kubectl get crds/prometheuses.monitoring.coreos.com' 'kubectl get crds/alertmanagers.monitoring.coreos.com'
  'kubectl get crds/podmonitors.monitoring.coreos.com' 'kubectl get crds/prometheusrules.monitoring.coreos.com'
  'kubectl get crds/servicemonitors.monitoring.coreos.com')
  for CMD in "${CMDs_ARR[@]}"
  do
    ATTEMPTS=0
    until $CMD || [ $ATTEMPTS -eq 60 ]
    do
      $CMD
      ATTEMPTS=$((ATTEMPTS + 1))
      sleep 5
    done
    if [[ $ATTEMPTS -eq 60 ]]; then
      echo "crds were not created successfully"
      exit 1
    fi
  done

  ATTEMPTS=0
  until [ $ATTEMPTS -eq 60 ]
  do
    $CMD
    ATTEMPTS=$((ATTEMPTS + 1))
    sleep 5
  done


  helm template --namespace istio-prometheus "${INSTALLER_DIR}"/istio-telemetry/prometheus-operator/ -f "${INSTALLER_DIR}"/global.yaml | kubectl apply -n istio-prometheus -f -

  # Install Promethues
  kubectl apply -f "${WD}/../prometheus-install/prometheus.yaml"
}

# shellcheck disable=SC2086
install_prometheus $1

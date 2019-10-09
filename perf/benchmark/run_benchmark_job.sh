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

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

# This is config for postsubmit cluster.
export VALUES="${VALUES:-values-istio-postsubmit.yaml}"
export DNS_DOMAIN="fake-dns.org"
export FORTIO_CLIENT_URL=""
export LC_ALL=C.UTF-8
export LANG=C.UTF-8

function setup_metrics() {
  # shellcheck disable=SC2155
  export FORTIO_CLIENT_URL=http://$(kubectl get services -n twopods fortioclient -o jsonpath="{.status.loadBalancer.ingress[0].ip}"):8080
  if [[ -z "$FORTIO_CLIENT_URL" ]];then
    kubectl -n twopods port-forward svc/fortioclient 8080:8080 &
    export FORTIO_CLIENT_URL=http://localhost:8080
  fi
  export PROMETHEUS_URL=http://localhost:9090
  kubectl -n istio-prometheus port-forward svc/istio-prometheus 9090:9090 &>/dev/null &
}

function collect_metrics() {
  local GENERATE_GRAPH=$1
  local PLOT_METRIC=$2
  CSV_OUTPUT="$(mktemp /tmp/benchmark_XXXX.csv)"
  pipenv install
  pipenv run python3 fortio.py $FORTIO_CLIENT_URL --csv_output="$CSV_OUTPUT" --prometheus=$PROMETHEUS_URL \
   --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,\
mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,\
mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy

  if [[ "$GENERATE_GRAPH" = true ]];then
    BENCHMARK_GRAPH="$(mktemp /tmp/benchmark_graph_XXXX.html)"
    pipenv run python3 graph.py "${CSV_OUTPUT}" "${PLOT_METRIC}" --charts_output="${BENCHMARK_GRAPH}"
    dt=$(date +'%Y%m%d-%H')
    RELEASE="$(cut -d'/' -f3 <<<"${CB_GCS_FULL_STAGING_PATH}")"
    GRAPH_NAME="${RELEASE}.${dt}.${PLOT_METRIC}"
    gsutil -q cp "${BENCHMARK_GRAPH}" "gs://$CB_GCS_BUILD_PATH/${GRAPH_NAME}"
  fi
}

RELEASE_TYPE="dev"
TAG=$(curl "https://storage.googleapis.com/istio-build/dev/latest")
echo "Setup istio release: $TAG"
pushd "${ROOT}/istio-install"
   ./setup_istio_release.sh "${TAG}" "${RELEASE_TYPE}"
popd
# install dependencies
cd "${WD}/runner"
pipenv install
# setup test
pushd "${WD}"
./setup_test.sh
popd

setup_metrics
echo "Start running perf benchmark test."
# For adding or modifying configurations, refer to perf/benchmark/README.md
# Configuration1
EXTRA_ARGS="--serversidecar --baseline"
CONN=16
QPS=500,1000,1500,2000
DURATION=300
METRIC="cpu"
# shellcheck disable=SC2086
pipenv run python3 runner.py ${CONN} ${QPS} ${DURATION} ${EXTRA_ARGS}
collect_metrics true ${METRIC}
METRIC="mem"
collect_metrics true ${METRIC}

# Configuration2
CONN=1,2,4,8,16,32,64
QPS=1000
METRIC="p90"
# shellcheck disable=SC2086
pipenv run python3 runner.py ${CONN} ${QPS} ${DURATION} ${EXTRA_ARGS}
collect_metrics true ${METRIC}

#TODO: Add more configurations, e.g. no mixer vs mixer comparison.

echo "perf benchmark test is done."
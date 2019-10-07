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
  # shellcheck disable=SC2155
  export CSV_OUTPUT="$(mktemp /tmp/benchmark_XXXX.csv)"
  pipenv install
  pipenv run python3 fortio.py $FORTIO_CLIENT_URL --csv_output="$CSV_OUTPUT" --prometheus=$PROMETHEUS_URL \
   --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,\
mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,\
mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy
}

function generate_graph() {
  local PLOT_METRIC=$1
  BENCHMARK_GRAPH="$(mktemp /tmp/benchmark_graph_XXXX.html)"
  pipenv run python3 graph.py "${CSV_OUTPUT}" "${PLOT_METRIC}" --charts_output="${BENCHMARK_GRAPH}"
  dt=$(date +'%Y%m%d-%H')
  RELEASE="$(cut -d'/' -f3 <<<"${CB_GCS_FULL_STAGING_PATH}")"
  GRAPH_NAME="${RELEASE}.${dt}.${PLOT_METRIC}"
  gsutil -q cp "${BENCHMARK_GRAPH}" "gs://$CB_GCS_BUILD_PATH/${GRAPH_NAME}"
}

function get_benchmark_data() {
  # shellcheck disable=SC2086
  pipenv run python3 runner.py ${CONN} ${QPS} ${DURATION} ${EXTRA_ARGS} ${MIXER_MODE}
  collect_metrics
  for metric in "${METRICS[@]}"
  do
    generate_graph "${metric}"
  done
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

# Setup fortio and prometheus
setup_metrics

echo "Start running perf benchmark test."
# For adding or modifying configurations, refer to perf/benchmark/README.md
EXTRA_ARGS="--serversidecar --baseline"
# Configuration Set1: CPU and memory with mixer enabled
MIXER_MODE="--mixer_mode mixer"
CONN=16
QPS=10,100,500,1000,2000,3000
DURATION=240
METRICS=(cpu mem)
get_benchmark_data

# Configuration Set2: Latency Quantiles with mixer enabled
CONN=1,2,4,8,16,32,64
QPS=1000
METRICS=(p50 p90 p99)
get_benchmark_data

# Configuration Set3: CPU and memory with mixer disabled
kubectl -n istio-system get cm istio -o yaml > /tmp/meshconfig.yaml
pipenv run python3 ./update_mesh_config.py disable_mixer /tmp/meshconfig.yaml | kubectl -n istio-system apply -f /tmp/meshconfig.yaml
MIXER_MODE="--mixer_mode nomixer"
CONN=16
QPS=10,100,500,1000,2000,3000
DURATION=240
METRICS=(cpu mem)
get_benchmark_data

# Configuration Set4: Latency Quantiles with mixer disabled
CONN=1,2,4,8,16,32,64
QPS=1000
METRICS=(p50 p90 p99)
get_benchmark_data

# Configuration Set5: CPU and memory with mixer v2
kubectl -n istio-system apply -f https://raw.githubusercontent.com/istio/proxy/master/extensions/stats/testdata/istio/metadata-exchange_filter.yaml
kubectl -n istio-system apply -f https://raw.githubusercontent.com/istio/proxy/master/extensions/stats/testdata/istio/stats_filter.yaml
MIXER_MODE="--mixer_mode mixerv2"
CONN=16
QPS=10,100,500,1000,2000,3000
DURATION=240
METRICS=(cpu mem)
get_benchmark_data

# Configuration Set6: Latency Quantiles with mixer v2
CONN=1,2,4,8,16,32,64
QPS=1000
METRICS=(p50 p90 p99)
get_benchmark_data

# TODO: Configuration Set5: Flame Graphs

echo "perf benchmark test is done."
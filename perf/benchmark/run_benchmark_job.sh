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

#!/bin/bash
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

function collect_metrics() {
  local GENERATE_GRAPH=$1

  FORTIO_CLIENT_URL=http://$(kubectl get services -n twopods fortioclient -o jsonpath="{.status.loadBalancer.ingress[0].ip}"):8080
  if [[ -z "$FORTIO_CLIENT_URL" ]];then
    kubectl -n twopods port-forward svc/fortioclient 8080:8080 &
    export FORTIO_CLIENT_URL=http://localhost:8080
  fi
  export PROMETHEUS_URL=http://localhost:9090
  kubectl -n istio-prometheus port-forward svc/istio-prometheus 9090:9090 &>/dev/null &

  CSV_OUTPUT="$(mktemp /tmp/benchmark_XXXX.csv)"
  pipenv install
  pipenv run python fortio.py $FORTIO_CLIENT_URL --csv_output="$CSV_OUTPUT" --prometheus=$PROMETHEUS_URL \
   --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,\
mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,\
mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy

  if [[ "$GENERATE_GRAPH" = true ]];then
    METRICS=(p99 mem cpu)
    for metric in "${METRICS[@]}"
    do
      BENCHMARK_GRAPH="$(mktemp /tmp/benchmark_graph_XXXX.html)"
      pipenv run python graph.py "${CSV_OUTPUT}" "${metric}" --charts_output="${BENCHMARK_GRAPH}"
      dt=$(date +'%Y%m%d-%H')
      RELEASE="$(cut -d'/' -f3 <<<"${CB_GCS_FULL_STAGING_PATH}")"
      GRAPH_NAME="${RELEASE}.${dt}"
      gsutil -q cp "${BENCHMARK_GRAPH}" "gs://$CB_GCS_BUILD_PATH/${GRAPH_NAME}"
    done
  fi
}

echo "Setup istio release: $TAG"
pushd "${ROOT}/istio-install"
  ./setup_istio.sh "${TAG}"
popd

cd "${WD}/runner"
pipenv install

EXTRA_ARGS="--serversidecar --baseline"
CONN=16,64
QPS=500,1000
DURATION=300
pushd "${WD}"
./setup_test.sh
popd

echo "Start running perf benchmark test."
# run towards all combinations of the parameter
# shellcheck disable=SC2086
pipenv run python runner.py ${CONN} ${QPS} ${DURATION} ${EXTRA_ARGS}
collect_metrics true

echo "perf benchmark test is done."
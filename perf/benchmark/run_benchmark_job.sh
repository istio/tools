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
export GCS_BUCKET="istio-build/perf"

function setup_metrics() {
  # shellcheck disable=SC2155
  INGRESS_IP="$(kubectl get services -n twopods fortioclient -o jsonpath="{.status.loadBalancer.ingress[0].ip}")"
  export FORTIO_CLIENT_URL=http://${INGRESS_IP}:8080
  if [[ -z "$INGRESS_IP" ]];then
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
  pipenv run python3 fortio.py ${FORTIO_CLIENT_URL} --csv_output="$CSV_OUTPUT" --prometheus=${PROMETHEUS_URL} \
   --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,\
mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,\
mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy

  gsutil -q cp "${CSV_OUTPUT}" "gs://${GCS_BUCKET}/${OUTPUT_DIR}/benchmark.csv"
}

function collect_flame_graph() {
    cd "${WD}/flame/flameoutput/"
    # shellcheck disable=SC2044
    # shellcheck disable=SC2006
    for i in `find . -name "*.svg"`
    do
      gsutil -q cp "${i}" "gs://${GCS_BUCKET}/${OUTPUT_DIR}/flamegraphs"
    done
}

function generate_graph() {
  local PLOT_METRIC=$1
  pipenv run python3 graph.py "${CSV_OUTPUT}" "${PLOT_METRIC}" --charts_output_dir="${LOCAL_OUTPUT_DIR}"
}

function get_benchmark_data() {
  # shellcheck disable=SC2086
  pipenv run python3 runner.py ${CONN} ${QPS} ${DURATION} ${EXTRA_ARGS} ${FLAME_GRAGH_ARG} ${MIXER_MODE}
  collect_metrics

  if ${FLAME_GRAGH_ARG} = "--perf=true"; then
    collect_flame_graph
  fi

  for metric in "${METRICS[@]}"
  do
    generate_graph "${metric}"
  done
  gsutil -q cp -r "${LOCAL_OUTPUT_DIR}" "gs://$GCS_BUCKET/${OUTPUT_DIR}/graphs"
}

function exit_handling() {
  # copy raw data from fortio client pod
  kubectl --namespace twopods cp "${FORTIO_CLIENT_POD}":/var/lib/fortio /tmp/rawdata -c shell
  gsutil -q cp -r /tmp/rawdata "gs://${GCS_BUCKET}/${OUTPUT_DIR}/rawdata"
  # output information for debugging
  kubectl logs -n twopods "${FORTIO_CLIENT_POD}" -c captured || true
  kubectl top pods --containers -n twopods || true
  kubectl describe pods "${FORTIO_CLIENT_POD}" -n twopods || true
}

function enable_perf_record() {
  nodes=$(kubectl get nodes -o=jsonpath='{.items[*].metadata.name}')
  for node in $nodes
  do
    gcloud compute ssh --command "sudo sysctl kernel.perf_event_paranoid=-1;sudo sysctl kernel.kptr_restrict=0;exit" \
    --zone us-central1-f bootstrap@"$node"
  done
}

RELEASE_TYPE="dev"
BRANCH="latest"
if [ "${GIT_BRANCH}" != "master" ];then
  BRANCH_NUM=$(echo "$GIT_BRANCH" | cut -f2 -d-)
  BRANCH="${BRANCH_NUM}-dev"
fi
# different branch tag resides in dev release directory like /latest, /1.4-dev, /1.5-dev etc.
TAG=$(curl "https://storage.googleapis.com/istio-build/dev/${BRANCH}")
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
dt=$(date +'%Y%m%d-%H')
export OUTPUT_DIR="benchmark_data-${GIT_BRANCH}.${dt}.${GIT_SHA}"
LOCAL_OUTPUT_DIR="/tmp/${OUTPUT_DIR}"
mkdir -p "${LOCAL_OUTPUT_DIR}"

# Setup fortio and prometheus
setup_metrics
FORTIO_CLIENT_POD=$(kubectl get pods -n twopods | grep fortioclient | awk '{print $1}')
export FORTIO_CLIENT_POD
FORTIO_SERVER_POD=$(kubectl get pods -n twopods | grep fortioserver | awk '{print $1}')
export FORTIO_SERVER_POD

# add trap to copy raw data when exiting, also output logging information for debugging
trap exit_handling ERR
trap exit_handling EXIT

echo "Start running perf benchmark test, data would be saved to GCS bucket: ${GCS_BUCKET}/${OUTPUT_DIR}"
# For adding or modifying configurations, refer to perf/benchmark/README.md
EXTRA_ARGS="--serversidecar --baseline"

enable_perf_record

# Configuration Set1: CPU and memory with mixer enabled
FLAME_GRAGH_ARG="--perf=false"
MIXER_MODE="--mixer_mode mixer"
CONN=16
QPS=10,100,500,1000,2000,3000
DURATION=240
METRICS=(cpu mem)
get_benchmark_data

# Configuration Set2: Latency Quantiles with mixer enabled
FLAME_GRAGH_ARG="--perf=true"
CONN=1,2,4,8,16,32,64
QPS=1000
METRICS=(p50 p90 p99)
get_benchmark_data
# restart proxy after each group(two sets)
kubectl exec -it -n twopods "${FORTIO_CLIENT_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST
kubectl exec -it -n twopods "${FORTIO_SERVER_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST

# Configuration Set3: CPU and memory with mixer disabled
kubectl -n istio-system get cm istio -o yaml > /tmp/meshconfig.yaml
pipenv run python3 "${WD}"/update_mesh_config.py disable_mixer /tmp/meshconfig.yaml | kubectl -n istio-system apply -f -
FLAME_GRAGH_ARG="--perf=false"
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
# restart proxy after each group
kubectl exec -it -n twopods "${FORTIO_CLIENT_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST
kubectl exec -it -n twopods "${FORTIO_SERVER_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST

# Configuration Set5: CPU and memory with telemetryv2 using NullVM.
kubectl -n istio-system apply -f https://raw.githubusercontent.com/istio/istio/master/tests/integration/telemetry/stats/prometheus/testdata/metadata_exchange_filter.yaml
kubectl -n istio-system apply -f https://raw.githubusercontent.com/istio/istio/master/tests/integration/telemetry/stats/prometheus/testdata/stats_filter.yaml
MIXER_MODE="--mixer_mode telemetryv2-nullvm"
CONN=16
QPS=10,100,500,1000,2000,3000
DURATION=240
METRICS=(cpu mem)
get_benchmark_data

# Configuration Set6: Latency Quantiles with telemetry v2 using NullVM.
CONN=1,2,4,8,16,32,64
QPS=1000
METRICS=(p50 p90 p99)
get_benchmark_data
# restart proxy after each group
kubectl exec -it -n twopods "${FORTIO_CLIENT_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST
kubectl exec -it -n twopods "${FORTIO_SERVER_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST

echo "perf benchmark test is done."

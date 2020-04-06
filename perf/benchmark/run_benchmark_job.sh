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
# Check https://github.com/istio/test-infra/blob/master/boskos/configs.yaml
# for existing resources types
export RESOURCE_TYPE="${RESOURCE_TYPE:-gke-perf-preset}"
export OWNER="${OWNER:-perf-tests}"
export PILOT_CLUSTER="${PILOT_CLUSTER:-}"
export USE_MASON_RESOURCE="${USE_MASON_RESOURCE:-True}"
export CLEAN_CLUSTERS="${CLEAN_CLUSTERS:-True}"
export NAMESPACE=${NAMESPACE:-'twopods-istio'}
export PROMETHEUS_NAMESPACE=${PROMETHEUS_NAMESPACE:-'istio-system'}

function setup_metrics() {
  # shellcheck disable=SC2155
  INGRESS_IP="$(kubectl get services -n "${NAMESPACE}" fortioclient -o jsonpath="{.status.loadBalancer.ingress[0].ip}")"
  export FORTIO_CLIENT_URL=http://${INGRESS_IP}:8080
  if [[ -z "$INGRESS_IP" ]];then
    kubectl -n "${NAMESPACE}" port-forward svc/fortioclient 8080:8080 &
    export FORTIO_CLIENT_URL=http://localhost:8080
  fi
  export PROMETHEUS_URL=http://localhost:9090
  kubectl -n "${PROMETHEUS_NAMESPACE}" port-forward svc/prometheus 9090:9090 &>/dev/null &
}

function collect_metrics() {
  # shellcheck disable=SC2155
  export CSV_OUTPUT="$(mktemp /tmp/benchmark_XXXX.csv)"
  pipenv run python3 fortio.py ${FORTIO_CLIENT_URL} --csv_output="$CSV_OUTPUT" --prometheus=${PROMETHEUS_URL} \
   --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,\
mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,\
mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy

  gsutil -q cp "${CSV_OUTPUT}" "gs://${GCS_BUCKET}/${OUTPUT_DIR}/benchmark.csv"
}

function collect_flame_graph() {
    FLAME_OUTPUT_DIR="${WD}/flame/flameoutput/"
    gsutil -q cp -r "${FLAME_OUTPUT_DIR}" "gs://${GCS_BUCKET}/${OUTPUT_DIR}/flamegraphs"
}

function generate_graph() {
  local PLOT_METRIC=$1
  pipenv run python3 graph.py "${CSV_OUTPUT}" "${PLOT_METRIC}" --charts_output_dir="${LOCAL_OUTPUT_DIR}"
}

function get_benchmark_data() {
  pushd "${WD}/runner"
  CONFIG_FILE="${1}"
  pipenv run python3 runner.py --config_file "${CONFIG_FILE}"
  collect_metrics
  popd
}

function exit_handling() {
  # copy raw data from fortio client pod
  kubectl --namespace "${NAMESPACE}" cp "${FORTIO_CLIENT_POD}":/var/lib/fortio /tmp/rawdata -c shell
  gsutil -q cp -r /tmp/rawdata "gs://${GCS_BUCKET}/${OUTPUT_DIR}/rawdata"
  # output information for debugging
  kubectl logs -n "${NAMESPACE}" "${FORTIO_CLIENT_POD}" -c captured || true
  kubectl top pods --containers -n "${NAMESPACE}" || true
  kubectl describe pods "${FORTIO_CLIENT_POD}" -n "${NAMESPACE}" || true
}

function enable_perf_record() {
  nodes=$(kubectl get nodes -o=jsonpath='{.items[*].metadata.name}')

  for node in $nodes
  do
    gcloud compute ssh --command "sudo sysctl kernel.perf_event_paranoid=-1;sudo sysctl kernel.kptr_restrict=0;exit" \
    --zone us-central1-f bootstrap@"$node"
  done
}

function setup_fortio_and_prometheus() {
    setup_metrics
    FORTIO_CLIENT_POD=$(kubectl get pods -n "${NAMESPACE}" | grep fortioclient | awk '{print $1}')
    export FORTIO_CLIENT_POD
    FORTIO_SERVER_POD=$(kubectl get pods -n "${NAMESPACE}" | grep fortioserver | awk '{print $1}')
    export FORTIO_SERVER_POD
}

function collect_envoy_info() {
  CONFIG_NAME=${1}
  POD_NAME=${2}
  FILE_SUFFIX=${3}

  ENVOY_DUMP_NAME="${POD_NAME}_${CONFIG_NAME}_${FILE_SUFFIX}.yaml"
  kubectl exec -n "${NAMESPACE}" "${POD_NAME}" -c istio-proxy -- curl http://localhost:15000/"${FILE_SUFFIX}" > "${ENVOY_DUMP_NAME}"
  gsutil -q cp -r "${ENVOY_DUMP_NAME}" "gs://${GCS_BUCKET}/${OUTPUT_DIR}/${FILE_SUFFIX}/${ENVOY_DUMP_NAME}"
}

function collect_config_dump() {
  collect_envoy_info "${1}" "${FORTIO_CLIENT_POD}" "config_dump"
  collect_envoy_info "${1}" "${FORTIO_SERVER_POD}" "config_dump"
}

function collect_clusters_info() {
  collect_envoy_info "${1}" "${FORTIO_CLIENT_POD}" "clusters"
  collect_envoy_info "${1}" "${FORTIO_SERVER_POD}" "clusters"
}

# install pipenv
if [[ $(command -v pipenv) == "" ]];then
  apt-get update && apt-get -y install python3-pip
  pip3 install pipenv
fi

# setup cluster
helm init --client-only
# shellcheck disable=SC1090
source "${ROOT}/../bin/setup_cluster.sh"
setup_e2e_cluster

# setup release info
RELEASE_TYPE="dev"
BRANCH="latest"
if [ "${GIT_BRANCH}" != "master" ];then
  BRANCH_NUM=$(echo "$GIT_BRANCH" | cut -f2 -d-)
  BRANCH="${BRANCH_NUM}-dev"
fi

# different branch tag resides in dev release directory like /latest, /1.4-dev, /1.5-dev etc.
TAG=$(curl "https://storage.googleapis.com/istio-build/dev/${BRANCH}")
echo "Setup istio release: $TAG"
# TAG is of the form like "1.5-alpha.sha"
# shellcheck disable=SC2155
export GIT_SHA=$(echo "$TAG" | cut -f3 -d.)

pushd "${ROOT}/istio-install"
   export INSTALL_WITH_ISTIOCTL="true"
   ./setup_istio_release.sh "${TAG}" "${RELEASE_TYPE}"
popd

# install dependencies
cd "${WD}/runner"
pipenv install

# setup istio test
pushd "${WD}"
export ISTIO_INJECT="true"
./setup_test.sh
popd
dt=$(date +'%Y%m%d-%H')
SHA=$(git rev-parse --short "${GIT_SHA}")
export OUTPUT_DIR="benchmark_data-${GIT_BRANCH}.${dt}.${SHA}"
LOCAL_OUTPUT_DIR="/tmp/${OUTPUT_DIR}"
mkdir -p "${LOCAL_OUTPUT_DIR}"

setup_fortio_and_prometheus

# add trap to copy raw data when exiting, also output logging information for debugging
trap exit_handling ERR
trap exit_handling EXIT

echo "Start running perf benchmark test, data would be saved to GCS bucket: ${GCS_BUCKET}/${OUTPUT_DIR}"

# enable flame graph
# enable_perf_record

DEFAULT_CR_PATH="${ROOT}/istio-install/istioctl_profiles/default.yaml"
# For adding or modifying configurations, refer to perf/benchmark/README.md
CONFIG_DIR="${WD}/configs/istio"

for dir in "${CONFIG_DIR}"/*; do
    pushd "${dir}"
    # install istio with custom overlay
    if [[ -e "./installation.yaml" ]]; then
       extra_overlay="-f ${dir}/installation.yaml"
    fi
    pushd "${ROOT}/istio-install/tmp"
      ./istioctl manifest apply -f "${DEFAULT_CR_PATH}" "${extra_overlay}" --force --wait
    popd

    # custom pre run
    if [[ -e "./prerun.sh" ]]; then
       # shellcheck disable=SC1091
       source prerun.sh
    fi

    config_name=$(echo "${dir}" | awk -F'/' '{print $NF}')
    # collect config dump after prerun.sh and before test run, to verify test setup is correct
    collect_config_dump "${config_name}"

    # run test and get data
    if [[ -e "./cpu_mem.yaml" ]]; then
       get_benchmark_data "${dir}/cpu_mem.yaml"
    fi
    if [[ -e "./latency.yaml" ]]; then
       get_benchmark_data "${dir}/latency.yaml"
    fi

    # collect clusters info after test run and before cleanup script postrun.sh
    collect_clusters_info "${config_name}"

    # custom post run
    if [[ -e "./postrun.sh" ]]; then
       # shellcheck disable=SC1091
       source postrun.sh
    fi
    # TODO: can be added to shared_postrun.sh
    # restart proxy after each group
    kubectl exec -n "${NAMESPACE}" "${FORTIO_CLIENT_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST
    kubectl exec -n "${NAMESPACE}" "${FORTIO_SERVER_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST

    popd
done

#echo "collect flame graph ..."
#collect_flame_graph

echo "perf benchmark test for istio is done."
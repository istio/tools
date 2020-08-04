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

export LC_ALL=C.UTF-8
export LANG=C.UTF-8

# Boskos cluster related Env vars
# This is the config for postsubmit cluster.
export VALUES="${VALUES:-values-istio-postsubmit.yaml}"
# Check https://github.com/istio/test-infra/blob/master/boskos/configs.yaml
# for existing resources types
export RESOURCE_TYPE="${RESOURCE_TYPE:-gke-perf-preset}"
export PILOT_CLUSTER="${PILOT_CLUSTER:-}"
export USE_MASON_RESOURCE="${USE_MASON_RESOURCE:-True}"
export CLEAN_CLUSTERS="${CLEAN_CLUSTERS:-True}"
export OWNER="${OWNER:-perf-tests}"

# Istio performance test related Env vars
export NAMESPACE=${NAMESPACE:-'twopods-istio'}
export PROMETHEUS_NAMESPACE=${PROMETHEUS_NAMESPACE:-'istio-system'}
export ISTIO_INJECT=${ISTIO_INJECT:-true}
export DNS_DOMAIN="fake-dns.org"
export LOAD_GEN_TYPE=${LOAD_GEN_TYPE:-"fortio"}
export FORTIO_CLIENT_URL=""

# Other Env vars
export GCS_BUCKET="istio-build/perf"
export TRIALRUN=${TRIALRUN:-"False"}


CLEANUP_PIDS=()

# Step 1: setup cluster
helm init --client-only
# shellcheck disable=SC1090
source "${ROOT}/../bin/setup_cluster.sh"
setup_e2e_cluster

# Step 2: install Istio
# Setup release info
BRANCH="latest"
if [[ "${GIT_BRANCH}" != "master" ]];then
  BRANCH_NUM=$(echo "$GIT_BRANCH" | cut -f2 -d-)
  BRANCH="${BRANCH_NUM}-dev"
fi

# Different branch tag resides in dev release directory like /latest, /1.4-dev, /1.5-dev etc.
INSTALL_VERSION=$(curl "https://storage.googleapis.com/istio-build/dev/${BRANCH}")
echo "Setup istio release: ${INSTALL_VERSION}"

pushd "${ROOT}/istio-install"
   DEV_VERSION=${INSTALL_VERSION} ./setup_istio.sh -f istioctl_profiles/default-overlay.yaml
popd

# Step 3: setup Istio performance test
pushd "${WD}"
export ISTIO_INJECT="true"
./setup_test.sh
popd

# Step 4: install Python dependencies
# Install pipenv
if [[ $(command -v pipenv) == "" ]];then
  apt-get update && apt-get -y install python3-pip
  pip3 install pipenv
fi

# Install dependencies
cd "${WD}/runner"
pipenv install

# Step 5: setup perf data local output directory
dt=$(date +'%Y%m%d')
# Current output dir should be like: 20200523_nighthawk_master_1.7-alpha.f19fb40b777e357b605e85c04fb871578592ad1e
export OUTPUT_DIR="${dt}_${LOAD_GEN_TYPE}_${GIT_BRANCH}_${INSTALL_VERSION}"
LOCAL_OUTPUT_DIR="/tmp/${OUTPUT_DIR}"
mkdir -p "${LOCAL_OUTPUT_DIR}"

# Step 6: setup fortio and prometheus
function setup_fortio_and_prometheus() {
    # shellcheck disable=SC2155
    INGRESS_IP="$(kubectl get services -n "${NAMESPACE}" fortioclient -o jsonpath="{.status.loadBalancer.ingress[0].ip}")"
    local report_port="8080"
    if [[ "${LOAD_GEN_TYPE}" == "nighthawk" ]]; then
        report_port="9076"
    fi

    export FORTIO_CLIENT_URL=http://${INGRESS_IP}:${report_port}
    if [[ -z "$INGRESS_IP" ]];then
        kubectl -n "${NAMESPACE}" port-forward svc/fortioclient ${report_port}:${report_port} &
        CLEANUP_PIDS+=("$!")
        export FORTIO_CLIENT_URL=http://localhost:${report_port}
    fi

    export PROMETHEUS_URL=http://localhost:9090
    kubectl -n "${PROMETHEUS_NAMESPACE}" port-forward svc/prometheus 9090:9090 &>/dev/null &
    CLEANUP_PIDS+=("$!")

    FORTIO_CLIENT_POD=$(kubectl get pods -n "${NAMESPACE}" | grep fortioclient | awk '{print $1}')
    export FORTIO_CLIENT_POD
    FORTIO_SERVER_POD=$(kubectl get pods -n "${NAMESPACE}" | grep fortioserver | awk '{print $1}')
    export FORTIO_SERVER_POD
}

setup_fortio_and_prometheus

# Step 7: setup exit handling
function exit_handling() {
  for pid in "${CLEANUP_PIDS[@]}"; do
    kill "${pid}"
  done

  if [[ "${TRIALRUN}" == "True" ]]; then
     exit 0
  fi

  # Copy raw data from fortio client pod
  kubectl --namespace "${NAMESPACE}" cp "${FORTIO_CLIENT_POD}":/var/lib/fortio /tmp/rawdata -c shell
  gsutil -q cp -r /tmp/rawdata "gs://${GCS_BUCKET}/${OUTPUT_DIR}/rawdata"
  # output information for debugging
  kubectl logs -n "${NAMESPACE}" "${FORTIO_CLIENT_POD}" -c captured || true
  kubectl top pods --containers -n "${NAMESPACE}" || true
  kubectl describe pods "${FORTIO_CLIENT_POD}" -n "${NAMESPACE}" || true
}

# add trap to copy raw data when exiting, also output logging information for debugging
trap exit_handling ERR
trap exit_handling EXIT

# Step 8: run Istio performance test
# Helper functions
function collect_flame_graph() {
    FLAME_OUTPUT_DIR="${WD}/flame/flameoutput"
    gsutil -q cp -r "${FLAME_OUTPUT_DIR}/*.svg" "gs://${GCS_BUCKET}/${OUTPUT_DIR}/flamegraphs"
}

function collect_metrics() {
  # shellcheck disable=SC2155
  export CSV_OUTPUT="$(mktemp /tmp/benchmark_XXXX.csv)"
  pipenv run python3 fortio.py ${FORTIO_CLIENT_URL} --csv_output="$CSV_OUTPUT" --prometheus=${PROMETHEUS_URL} \
   --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_istio_proxy_fortioclient,\
cpu_mili_avg_istio_proxy_fortioserver,cpu_mili_avg_istio_proxy_istio-ingressgateway,mem_Mi_avg_istio_proxy_fortioclient,\
mem_Mi_avg_istio_proxy_fortioserver,mem_Mi_avg_istio_proxy_istio-ingressgateway

  gsutil -q cp "${CSV_OUTPUT}" "gs://${GCS_BUCKET}/${OUTPUT_DIR}/benchmark.csv"
}

function run_benchmark_test() {
  pushd "${WD}/runner"
  CONFIG_FILE="${1}"
  pipenv run python3 runner.py --config_file "${CONFIG_FILE}"

  if [[ "${TRIALRUN}" == "False" ]]; then
    collect_metrics
  fi
  popd
}

function read_perf_test_conf() {
  perf_test_conf="${1}"
  while IFS="=" read -r key value; do
    case "$key" in
      '#'*) ;;
      *)
        # shellcheck disable=SC2086
        export ${key}="${value}"
    esac
  done < "${perf_test_conf}"
}

function collect_envoy_info() {
  CONFIG_NAME=${1}
  POD_NAME=${2}
  FILE_SUFFIX=${3}

  ENVOY_DUMP_NAME="${LOAD_GEN_TYPE}_${POD_NAME}_${CONFIG_NAME}_${FILE_SUFFIX}.yaml"
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

function collect_pod_spec() {
  POD_NAME=${1}
  POD_SPEC_NAME="${LOAD_GEN_TYPE}_${POD_NAME}.yaml"
  if [[ "${POD_NAME}" == "prometheus"* ]]; then
    kubectl get pods "${POD_NAME}" -n "${PROMETHEUS_NAMESPACE}" -o yaml > "${POD_SPEC_NAME}"
  else
    kubectl get pods "${POD_NAME}" -n "${NAMESPACE}" -o yaml > "${POD_SPEC_NAME}"
  fi
  gsutil -q cp -r "${POD_SPEC_NAME}" "gs://${GCS_BUCKET}/${OUTPUT_DIR}/pod_spec/${POD_SPEC_NAME}"
}

# Start run perf test
echo "Start to run perf benchmark test, all collected data will be dumped to GCS bucket: ${GCS_BUCKET}/${OUTPUT_DIR}"

# For adding or modifying configurations, refer to perf/benchmark/README.md
CONFIG_DIR="${WD}/configs/istio"
# Read through perf test configuration file to determine which group of test configuration to run or not run
read_perf_test_conf "${WD}/configs/run_perf_test.conf"

for dir in "${CONFIG_DIR}"/*; do
    # Get the last directory name after splitting dir path by '/', which is the configuration dir name
    config_name="$(basename "${dir}")"
    # skip the test config which is disabled for running
    if ! ${!config_name:-false}; then
        continue
    fi

    pushd "${dir}"

    # Install istio with custom overlay
    if [[ -e "./installation.yaml" ]]; then
       extra_overlay="-f ${dir}/installation.yaml"
    fi
    pushd "${ROOT}/istio-install"
      DEV_VERSION=${INSTALL_VERSION} ./setup_istio.sh -f istioctl_profiles/default-overlay.yaml "${extra_overlay}"
    popd

    # Custom pre-run
    if [[ -e "./prerun.sh" ]]; then
       # shellcheck disable=SC1091
       source prerun.sh
    fi

    # TRIALRUN as a pre-submit check, only run agaist the first set of enabled perf run in the perf_conf file
    if [[ "${TRIALRUN}" == "True" ]]; then
       run_benchmark_test "${WD}/configs/trialrun.yaml"
       collect_flame_graph
       break
    fi

    # Collect config_dump after prerun.sh and before test run, in order to verify test setup is correct
    collect_config_dump "${config_name}"

    # Collect pod spec
    collect_pod_spec "${FORTIO_CLIENT_POD}"
    collect_pod_spec "${FORTIO_SERVER_POD}"

    # Collect prometheus pod spec
    PROM_POD=$(kubectl get pods -n "${PROMETHEUS_NAMESPACE}" | grep prometheus | awk '{print $1}')
    collect_pod_spec "${PROM_POD}"

    # Run test and collect data
    if [[ -e "./cpu_mem.yaml" ]]; then
       run_benchmark_test "${dir}/cpu_mem.yaml"
    fi

    if [[ -e "./latency.yaml" ]]; then
       run_benchmark_test "${dir}/latency.yaml"
    fi

    # Collect clusters info after test run and before cleanup postrun.sh run
    collect_clusters_info "${config_name}"

    # Custom post run
    if [[ -e "./postrun.sh" ]]; then
       # shellcheck disable=SC1091
       source postrun.sh
    fi

    collect_flame_graph
    # TODO: can be added to shared_postrun.sh

    # restart proxy after each group
    kubectl exec -n "${NAMESPACE}" "${FORTIO_CLIENT_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST
    kubectl exec -n "${NAMESPACE}" "${FORTIO_SERVER_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST

    popd
done

echo "Istio performance benchmark test is done!"

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
export NAMESPACE="${NAMESPACE:-twopods}"
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
  pipenv install
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
  CONFIG_FILE="${1}"
  pipenv run python3 runner.py --config_file "${CONFIG_FILE}"
  collect_metrics
#  TODO: replace with new graph generation code
#  for metric in "${METRICS[@]}"
#  do
#    generate_graph "${metric}"
#  done
#  gsutil -q cp -r "${LOCAL_OUTPUT_DIR}" "gs://$GCS_BUCKET/${OUTPUT_DIR}/graphs"
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

# Setup fortio and prometheus
function setup_fortio_and_prometheus() {
    setup_metrics
    FORTIO_CLIENT_POD=$(kubectl get pods -n "${NAMESPACE}" | grep fortioclient | awk '{print $1}')
    export FORTIO_CLIENT_POD
    FORTIO_SERVER_POD=$(kubectl get pods -n "${NAMESPACE}" | grep fortioserver | awk '{print $1}')
    export FORTIO_SERVER_POD
}

function prerun_v2_nullvm() {
  kubectl -n istio-system apply -f https://raw.githubusercontent.com/istio/istio/"${GIT_BRANCH}"/tests/integration/telemetry/stats/prometheus/testdata/metadata_exchange_filter.yaml
  kubectl -n istio-system apply -f https://raw.githubusercontent.com/istio/istio/"${GIT_BRANCH}"/tests/integration/telemetry/stats/prometheus/testdata/stats_filter.yaml
}

function prerun_nomixer() {
  kubectl -n istio-system get cm istio -o yaml > /tmp/meshconfig.yaml
  pipenv run python3 "${WD}"/update_mesh_config.py disable_mixer /tmp/meshconfig.yaml | kubectl -n istio-system apply -f -
}

# Explicitly create meshpolicy to ensure the test is running as plaintext.
function prerun_plaintext() {
  echo "Saving current mTLS config first"
  kubectl -n "${NAMESPACE}"  get dr -oyaml > "${LOCAL_OUTPUT_DIR}/destionation-rule.yaml"
  kubectl -n "${NAMESPACE}"  get policy -oyaml > "${LOCAL_OUTPUT_DIR}/authn-policy.yaml"
  echo "Deleting Authn Policy and DestinationRule"
  kubectl -n "${NAMESPACE}" delete dr --all
  kubectl -n "${NAMESPACE}" delete policy --all
  echo "Configure plaintext..."
  cat <<EOF | kubectl apply -f -
apiVersion: "authentication.istio.io/v1alpha1"
kind: "Policy"
metadata:
  name: "default"
  namespace: "${NAMESPACE}"
spec: {}
EOF
  # Explicitly disable mTLS by DestinationRule to avoid potential auto mTLS effect.
  cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: plaintext-dr-twopods
  namespace: ${NAMESPACE}
spec:
  host:  "*.svc.${NAMESPACE}.cluster.local"
  trafficPolicy:
    tls:
      mode: DISABLE
EOF
}

function postrun_plaintext() {
  echo "Delete the plaintext related config..."
  kubectl delete policy -n"${NAMESPACE}" default
  kubectl delete DestinationRule -n"${NAMESPACE}" plaintext-dr-twopods
  echo "Restoring original Authn Policy and DestinationRule config..."
  kubectl apply -f "${LOCAL_OUTPUT_DIR}/authn-policy.yaml"
  kubectl apply -f "${LOCAL_OUTPUT_DIR}/destionation-rule.yaml"
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
pushd "${ROOT}/istio-install"
   export INSTALL_WITH_ISTIOCTL="true"
   ./setup_istio_release.sh "${TAG}" "${RELEASE_TYPE}"
popd

# install dependencies
cd "${WD}/runner"
pipenv install

# setup test
pushd "${WD}"
export ISTIO_INJECT="true"
./setup_test.sh
popd
dt=$(date +'%Y%m%d-%H')
export OUTPUT_DIR="benchmark_data-${GIT_BRANCH}.${dt}.${GIT_SHA}"
LOCAL_OUTPUT_DIR="/tmp/${OUTPUT_DIR}"
mkdir -p "${LOCAL_OUTPUT_DIR}"

setup_fortio_and_prometheus

# add trap to copy raw data when exiting, also output logging information for debugging
trap exit_handling ERR
trap exit_handling EXIT

echo "Start running perf benchmark test, data would be saved to GCS bucket: ${GCS_BUCKET}/${OUTPUT_DIR}"

# enable flame graph
enable_perf_record

# For adding or modifying configurations, refer to perf/benchmark/README.md
CONFIG_DIR="${WD}/configs"

for f in "${CONFIG_DIR}"/*; do
    fn=$(basename "${f}")
    # pre run
    if [[ "${fn}" =~ "none" ]];then
        prerun_nomixer
    elif [[ "${fn}" =~ "telemetryv2" ]];then
        prerun_v2_nullvm
    elif [[ "${fn}" =~ "plaintext" ]]; then
        prerun_plaintext
    fi

    get_benchmark_data "${f}"

    # post run

    # remove policy configured if any
    if [[ "${fn}" =~ "plaintext" ]]; then
      postrun_plaintext
    fi

    # restart proxy after each group
    kubectl exec -n "${NAMESPACE}" "${FORTIO_CLIENT_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST
    kubectl exec -n "${NAMESPACE}" "${FORTIO_SERVER_POD}" -c istio-proxy -- curl http://localhost:15000/quitquitquit -X POST
    
done

echo "collect flame graph ..."
collect_flame_graph

echo "perf benchmark test for istio is done."

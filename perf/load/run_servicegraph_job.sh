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

# Check https://github.com/istio/test-infra/blob/master/boskos/configs.yaml
# for existing resources types
export RESOURCE_TYPE="${RESOURCE_TYPE:-gke-perf-preset}"
export OWNER="${OWNER:-perf-tests}"
export PILOT_CLUSTER="${PILOT_CLUSTER:-}"
export USE_MASON_RESOURCE="${USE_MASON_RESOURCE:-True}"
export CLEAN_CLUSTERS="${CLEAN_CLUSTERS:-True}"
# TODO: use operator installation profile
export VALUES="${VALUES:-values-istio-postsubmit.yaml}"
export DNS_DOMAIN="fake-dns.org"
export CMD=""
export DELETE=""
export LC_ALL=C.UTF-8
export LANG=C.UTF-8

helm init --client-only
# shellcheck disable=SC1090
source "${ROOT}/../bin/setup_cluster.sh"
setup_e2e_cluster
export GCS_BUCKET="istio-build/perf"
dt=$(date +'%Y%m%d')
SHA=$(git rev-parse --short "${GIT_SHA}")
export OUTPUT_DIR="load-${GIT_BRANCH}.${dt}.${SHA}"

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
   ./setup_istio_release.sh "${TAG}" "${RELEASE_TYPE}"
popd

function service_graph() {
  # shellcheck disable=SC1090
  source "${WD}/common.sh"
  NAMESPACE_NUM=20
  START_NUM=0
  start_servicegraphs "${NAMESPACE_NUM}" "${START_NUM}"

  # Run the test for some time
  echo "Run the test for ${TIME_TO_RUN_PERF_TESTS} seconds"
  pod=$(kubectl get pod --namespace istio-system --selector="app=prometheus" --output jsonpath='{.items[0].metadata.name}')
  kubectl -n istio-system port-forward "$pod" 8060:9090 > /tmp/forward &

  sleep 5s
}

function export_metrics() {
  if [[ $(command -v pipenv) == "" ]];then
    apt-get update && apt-get -y install python3-pip
    pip3 install pipenv
  fi
  OUTPUT_PATH=${OUTPUT_PATH:-"/tmp/output"}

  mkdir -p "${OUTPUT_PATH}"

  load_metrics="${OUTPUT_PATH}/load_metrics.txt"
  rm "${load_metrics}" || true

  pushd "${ROOT}/benchmark/runner"
  pipenv install
  count="$((TIME_TO_RUN_PERF_TESTS / 60))"
    echo "Get metric $count time(s)."
    for i in $(seq 1 "$count");
    do
      echo "Running for $i min"
      sleep 1m
      pipenv run python3 prom.py http://localhost:8060 60 --no-aggregate >> "${load_metrics}"
    done

    gsutil -q cp "${load_metrics}" "gs://$CB_GCS_BUILD_PATH/load_metrics.txt"

  popd
}

echo "Start running service graph load test."
export TIME_TO_RUN_PERF_TESTS=${TIME_TO_RUN_PERF_TESTS:-3000}
service_graph
export_metrics
echo "Service graph load test is done."
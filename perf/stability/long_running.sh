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

set -eux

# envs for Istio installation
export TAG="${TAG:-}"
export VERSION="${VERSION:-}"
export RELEASE_URL="${RELEASE_URL:-}"
export DNS_DOMAIN="${DNS_DOMAIN:-release-qual.qualistio.org}"
export LOCAL_ISTIO_PATH="${LOCAL_ISTIO_PATH:-}"
export NAMESPACE_NUM="${NAMESPACE_NUM:-15}"
export SKIP_ISTIO_SETUP="${SKIP_ISTIO_SETUP:-false}"
export CANARY_UPGRADE_MODE="${CANARY_UPGRADE_MODE:-false}"
export PROMETHEUS_NAMESPACE="${PROMETHEUS_NAMESPACE:-istio-prometheus}"

# shellcheck disable=SC2153
if [[ -z "${PROJECT_ID:-}" ]] || [[ -z "${CLUSTER_NAME:-}" ]] ;then
  echo "You need to set PROJECT_ID and CLUSTER_NAME for where the test would be running"
  exit 1
fi

# setup Istio
if [[ ${SKIP_ISTIO_SETUP} != "true" ]];then
  pushd "${ROOT}/istio-install"
  # shellcheck disable=SC2199
  if [[ -z "${@-}" ]];then
    ./setup_istio.sh
  else
    ./setup_istio.sh "${@}"
  fi
  popd
fi

export NOT_INJECTED="True"
BRANCH="${TAG}"
if [[ -z "${BRANCH}" ]];then
  if [[ -n "${VERSION}" ]];then
    BRANCH="${VERSION}"
  elif [[ -n "${RELEASE_URL}" ]];then
    BRANCH="${RELEASE_URL}"
  fi
fi

NAMESPACE="istio-prometheus" ./setup_test.sh alertmanager
kubectl apply -f ./alertmanager/prometheusrule.yaml

# Setup workloads
pushd "${ROOT}/load"
# shellcheck disable=SC1091
source "./common.sh"
START_NUM=0
export START_NUM="${START_NUM:-0}"
export DELETE=""
export CMD=""
export WD="${ROOT}/load"
if [[ -z "${MULTI_CLUSTER-}" ]];then
  start_servicegraphs "${NAMESPACE_NUM}" "${START_NUM}"
else
 # run on two cluster
 CTX1=${CTX1} CTX=${CTX2} start_servicegraphs_multicluster "${NAMESPACE_NUM}" "${START_NUM}"
fi
popd

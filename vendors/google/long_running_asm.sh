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

# envs for spanner connection
export PROJECT_ID="${PROJECT_ID:-}"
export CLUSTER_NAME="${CLUSTER_NAME:-}"
export INSTANCE="${INSTANCE:-release-qual}"
export DBNAME="${DBNAME:-main}"
export MS_TABLE_NAME="${MS_TABLE_NAME:-MonitorStatus}"
export TESTID="${TESTID:-default}"

# below envs are for ASM installation
export TAG="${TAG:-}"
export VERSION="${VERSION:-}"
export RELEASE_URL="${RELEASE_URL:-}"
export NAMESPACE_NUM="${NAMESPACE_NUM:-15}"
export INSTALL_ASM="${INSTALL_ASM:-}"
export PROJECT_ID="${PROJECT_ID:-}"
export CLUSTER_NAME="${CLUSTER_NAME:-}"
export CLUSTER_LOCATION="${CLUSTER_LOCATION:-}"
export RELEASE="${RELEASE:-}"
export MULTI_CLUSTER="${MULTI_CLUSTER:-}"
export CTX1="${CTX1:-}"
export CTX2="${CTX2:-}"

export NOT_INJECTED="True"
BRANCH="${TAG}"
if [[ -z "${BRANCH}" ]];then
  if [[ -n "${VERSION}" ]];then
    BRANCH="${VERSION}"
  elif [[ -n "${RELEASE_URL}" ]];then
    BRANCH="${RELEASE_URL}"
  fi
fi

# deploy alertmanager related resources
DT=$(date +'%Y%m%d%H')
TESTID="${DT}-${BRANCH}"
HELM_ARGS="--set values.projectID=${PROJECT_ID} --set values.clusterName=${CLUSTER_NAME} --set values.branch=${BRANCH} --set values.instance=${INSTANCE} --set values.dbName=${DBNAME} --set values.testID=${TESTID} --set values.msTableName=${MS_TABLE_NAME}"
NAMESPACE="istio-prometheus" "${WD}"/setup_test.sh alertmanager "${HELM_ARGS}"
kubectl apply -f "${WD}/alertmanager/prometheusrule.yaml"

# deploy log scanner
kubectl create ns logs-checker || true
kubectl create configmap logs-checker --from-file=./logs-checker/check_k8s_logs.sh -n logs-checker || true
"${WD}"/setup_test.sh logs-checker

# This part would be only needed when we run the fully automated jobs on a dedicated cluster
# It would upgrade control plane and data plane to newer dev release every 48h.
# deploy canary upgrader
kubectl create ns canary-upgrader || true
kubectl create configmap canary-script --from-file=./canary-upgrader/canary_upgrade.sh --from-file=./../istio-install/setup_istio.sh -n canary-upgrader || true
"${WD}"/setup_test.sh canary-upgrader

# Setup workloads
# shellcheck disable=SC1091
source "./setup_workload.sh"
export START_NUM="${START_NUM:-0}"
export DELETE=""
export CMD=""
export WD="${ROOT}/load"
if [[ -z ${MULTI_CLUSTER} ]];then
  start_servicegraphs "${NAMESPACE_NUM}" "${START_NUM}"
else
 # run on two cluster
 CTX1=${CTX1} CTX=${CTX2} start_servicegraphs_multicluster "${NAMESPACE_NUM}" "${START_NUM}"
fi

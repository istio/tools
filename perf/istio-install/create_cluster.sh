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

# set -x
# Creates a standard GKE cluster for testing.

# shellcheck disable=SC2086
WD=$(dirname $0)
# shellcheck disable=SC2086
WD=$(cd $WD || exit; pwd)

# get default GKE cluster version for zone
function default_gke_version() {
  local zone=${1:?"zone is required"}
  # shellcheck disable=SC2155
  local temp_fname=$(mktemp)

  # shellcheck disable=SC2086
  gcloud container get-server-config --zone "${zone}"  > ${temp_fname} 2>&1
  # shellcheck disable=SC2181
  if [[ $? -ne 0 ]];then
    cat "${temp_fname}"
    exit 1
  fi

  # shellcheck disable=SC2002
  gke_ver=$(cat "${temp_fname}" | grep defaultClusterVersion | awk '{print $2}')
  echo "${gke_ver}"
  rm -rf "${temp_fname}"
}

# Required params
PROJECT_ID=${PROJECT_ID:?"project id is required"}
CLUSTER_NAME=${1:?"cluster name is required"}


# Optional params
ZONE=${ZONE:-us-central1-a}
# specify REGION to create a regional cluster

# Specify GCP_SA to create and use a specific service account.
# GCP_SA

# Sizing
DISK_SIZE=${DISK_SIZE:-64}
MACHINE_TYPE=${MACHINE_TYPE:-n1-standard-32}
MIN_NODES=${MIN_NODES:-"4"}
MAX_NODES=${MAX_NODES:-"70"}

MAXPODS_PER_NODE=100

# Labels and version
ISTIO_VERSION=${ISTIO_VERSION:?"Istio version label is required"}

if [[ -n "${GCP_SA}" ]];then
  "${WD}/create_sa.sh" "${GCP_SA}"
fi

DEFAULT_GKE_VERSION=$(default_gke_version "${ZONE}")
# shellcheck disable=SC2181
if [[ $? -ne 0 ]];then
  echo "${DEFAULT_GKE_VERSION}"
  exit 1
fi

GKE_VERSION=${GKE_VERSION-${DEFAULT_GKE_VERSION}}

# default scope for reference
# shellcheck disable=SC2034
SCOPES_DEFAULT="https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/trace.append"

# Full scope is needed for the context graph API
SCOPES_FULL="https://www.googleapis.com/auth/cloud-platform"

SCOPES="${SCOPES_FULL}"

# A label cannot have "." in it, replace "." with "_"
# shellcheck disable=SC2001
ISTIO_VERSION=$(echo "${ISTIO_VERSION}" | sed 's/\./_/g')

function gc() {
  # shellcheck disable=SC2236
  if [[ -n "${REGION}" ]];then
    ZZ="--region ${REGION}"
  else
    ZZ="--zone ${ZONE}"
  fi

  SA=""
  # shellcheck disable=SC2236
  if [[ -n "${GCP_SA}" ]];then
    SA="--identity-namespace=${PROJECT_ID}.svc.id.goog --service-account=${GCP_SA}@${PROJECT_ID}.iam.gserviceaccount.com --workload-metadata-from-node=EXPOSED"
  fi

  # shellcheck disable=SC2048
  # shellcheck disable=SC2086
  echo gcloud $* ${ZZ} ${SA}

  # shellcheck disable=SC2236
  if [[ -n "${DRY_RUN}" ]];then
    return
  fi

  # shellcheck disable=SC2086
  # shellcheck disable=SC2048
  gcloud $* ${ZZ} ${SA}
}

NETWORK_SUBNET="--create-subnetwork name=${CLUSTER_NAME}-subnet"
# shellcheck disable=SC2236
if [[ -n "${USE_SUBNET}" ]];then
  NETWORK_SUBNET="--network ${USE_SUBNET}"
fi

ADDONS="HorizontalPodAutoscaling,HttpLoadBalancing,KubernetesDashboard"
# shellcheck disable=SC2236
if [[ -n "${ISTIO_ADDON}" ]];then
  ADDONS+=",Istio"
fi
# shellcheck disable=SC2086
# shellcheck disable=SC2046
gc beta container \
  --project "${PROJECT_ID}" \
  clusters create "${CLUSTER_NAME}" \
  --no-enable-basic-auth --cluster-version "${GKE_VERSION}" \
  --machine-type "${MACHINE_TYPE}" --image-type "COS" --disk-type "pd-standard" --disk-size "${DISK_SIZE}" \
  --scopes "${SCOPES}" \
  --num-nodes "${MIN_NODES}" --enable-autoscaling --min-nodes "${MIN_NODES}" --max-nodes "${MAX_NODES}" --max-pods-per-node "${MAXPODS_PER_NODE}" \
  --enable-stackdriver-kubernetes \
  --enable-ip-alias \
  ${NETWORK_SUBNET} \
  --default-max-pods-per-node "${MAXPODS_PER_NODE}" \
  --addons "${ADDONS}" \
  --enable-network-policy --enable-autoupgrade --enable-autorepair --labels csm=1,test-date=$(date +%Y-%m-%d),version=${ISTIO_VERSION},operator=user_${USER}

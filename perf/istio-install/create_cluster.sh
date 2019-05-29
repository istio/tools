#!/bin/bash
# set -x
# Creates a standard cluster for testing.

# get default cluster version for zone
function default_cluster() {
  local zone=${1:?"zone required"}
  local temp_ver=$(mktemp)

  gcloud container get-server-config --zone "${zone}"  >${temp_ver} 2>&1
  if [[ $? -ne 0 ]];then
    cat "${temp_ver}"
    exit 1
  fi

  ver=$(cat "${temp_ver}" | grep defaultClusterVersion | awk '{print $2}')
  echo "${ver}"
  rm -Rf "${temp_ver}"
}

# Required params
PROJECT_ID=${PROJECT_ID:?"project id is required"}
CLUSTER_NAME=${1:?"cluster name"}

# Optional params
ZONE=${ZONE:-us-central1-a}

# Sizing
DISK_SIZE=${DISK_SIZE:-64}
MACHINE_TYPE=${MACHINE_TYPE:-n1-standard-32}
MIN_NODES=${MIN_NODES:-"4"}
MAX_NODES=${MAX_NODES:-"70"}

MAXPODS_PER_NODE=100

# Labels and version
ISTIO_VERSION=${ISTIO_VERSION:?"Istio version label"}
DEFAULT_GKE_VERSION=$(default_cluster "${ZONE}")
if [[ $? -ne 0 ]];then
  echo "${DEFAULT_GKE_VERSION}"
  exit 1
fi
GKE_VERSION=${GKE_VERSION-${DEFAULT_GKE_VERSION}}

# default scope for reference
SCOPES_DEFAULT="https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/trace.append"

# Full scope is needed for the context graph API
SCOPES_FULL="https://www.googleapis.com/auth/cloud-platform"

SCOPES="${SCOPES_FULL}"

# A label cannot have "." in it.
ISTIO_VERSION=$(echo ${ISTIO_VERSION} | sed 's/\./_/g')

function gc() {
  echo $*

  if [[ ! -z "${DRY_RUN}" ]];then
    return
  fi

  gcloud $*
}

gc beta container \
  --project "${PROJECT_ID}" \
  clusters create "${CLUSTER_NAME}" \
  --zone "${ZONE}" \
  --no-enable-basic-auth --cluster-version "${GKE_VERSION}" \
  --machine-type "${MACHINE_TYPE}" --image-type "COS" --disk-type "pd-standard" --disk-size "${DISK_SIZE}" \
  --scopes "${SCOPES}" \
  --num-nodes "${MIN_NODES}" --enable-autoscaling --min-nodes "${MIN_NODES}" --max-nodes "${MAX_NODES}" --max-pods-per-node "${MAXPODS_PER_NODE}" \
  --enable-stackdriver-kubernetes \
  --enable-ip-alias --create-subnetwork name="${CLUSTER_NAME}-subnet" \
  --default-max-pods-per-node "${MAXPODS_PER_NODE}" \
  --addons HorizontalPodAutoscaling,HttpLoadBalancing,KubernetesDashboard \
  --enable-network-policy --enable-autoupgrade --enable-autorepair --labels test-date=$(date +%Y-%m-%d),version=${ISTIO_VERSION},operator=user_${USER}

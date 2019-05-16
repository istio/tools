#!/bin/bash

# Creates a standard cluster for testing.
# You may adjust the initial size of the node pool

PROJECT_ID=${PROJECT_ID:?"project id is required"}
CLUSTER_NAME=${1:?"cluster name"}
ZONE=${ZONE:-us-central1-a}
DEFAULT_VERSION=$(gcloud container get-server-config --zone us-central1-a  2>/dev/null | grep defaultClusterVersion | awk '{print $2}')
GKE_VERSION=${GKE_VERSION-${DEFAULT_VERSION}}
SCOPES="https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/trace.append"
MAXPODS_PER_NODE=100
MIN_NODES=${MIN_NODES:-"4"}
MAX_NODES=${MAX_NODES:-"70"}

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
  --machine-type "n1-standard-32" --image-type "COS" --disk-type "pd-standard" --disk-size "64" \
  --scopes "${SCOPES}" \
  --max-pods-per-node "${MAXPODS_PER_NODE}" --num-nodes "1" --enable-stackdriver-kubernetes --enable-ip-alias --create-subnetwork name="${CLUSTER_NAME}-subnet" \
  --default-max-pods-per-node "${MAXPODS_PER_NODE}" \
  --addons HorizontalPodAutoscaling,HttpLoadBalancing,KubernetesDashboard \
  --enable-network-policy --enable-autoupgrade --enable-autorepair --labels test-date=$(date +%Y-%m-%d),version=v12,operator=user_${USER}

gc beta container \
  --project "${PROJECT_ID}" \
  node-pools create "service-graph-pool" --cluster "${CLUSTER_NAME}" --zone "${ZONE}" \
  --node-version "${GKE_VERSION}" \
  --machine-type "n1-standard-32" --image-type "COS" --disk-type "pd-standard" --disk-size "64" \
  --scopes "${SCOPES}" \
  --num-nodes "${MIN_NODES}" --enable-autoscaling --min-nodes "${MIN_NODES}" --max-nodes "${MAX_NODES}" --no-enable-autoupgrade --enable-autorepair --max-pods-per-node "${MAXPODS_PER_NODE}"

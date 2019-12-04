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

set -euo pipefail

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

set +u # Allow referencing unbound variable $CLUSTER
if [[ -z ${CLUSTER} ]]; then
  CLUSTER_NAME=${1:?"cluster name is required"}
else
  CLUSTER_NAME=${CLUSTER}
fi
set -u

# Optional params
ZONE=${ZONE:-us-central1-a}
# specify REGION to create a regional cluster

# Specify GCP_SA to create and use a specific service account.
# Default is to use same name as the cluster - each cluster can have different
# IAM permissions. This also enables workloadIdentity, which is recommended for GCP
GCP_SA=${GCP_SA:-$CLUSTER_NAME}
GCP_CTL_SA=${GCP_CTL_SA:-${CLUSTER_NAME}-control}

# Sizing
DISK_SIZE=${DISK_SIZE:-64}
MACHINE_TYPE=${MACHINE_TYPE:-n1-standard-32}
MIN_NODES=${MIN_NODES:-"4"}
MAX_NODES=${MAX_NODES:-"70"}
IMAGE=${IMAGE:-"COS"}
MAXPODS_PER_NODE=100

# Labels and version
ISTIO_VERSION=${ISTIO_VERSION:-master}

# Export CLUSTER_NAME so it will be set for the create_sa.sh script, which will
# create a google-cloud-key.json file in `./${CLUSTER_NAME}/`.
export CLUSTER_NAME
mkdir -p "${CLUSTER_NAME}"
"${WD}/create_sa.sh" "${GCP_SA}" "${GCP_CTL_SA}"

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

# Full scope is needed for the context graph API and NEG integration
SCOPES_FULL="https://www.googleapis.com/auth/cloud-platform"

SCOPES="${SCOPES_FULL}"



# A label cannot have "." in it, replace "." with "_"
# shellcheck disable=SC2001
ISTIO_VERSION=$(echo "${ISTIO_VERSION}" | sed 's/\./_/g')

function gc() {
  # shellcheck disable=SC2236
  if [[ -n "${REGION}" ]];then
    ZZ="--region=${REGION}"
  else
    ZZ="--zone=${ZONE}"
  fi

  SA=""
  # shellcheck disable=SC2236
  if [[ -n "${GCP_SA}" ]];then
    SA=("--identity-namespace=${PROJECT_ID}.svc.id.goog" "--service-account=${GCP_SA}@${PROJECT_ID}.iam.gserviceaccount.com" "--workload-metadata-from-node=EXPOSED")
  fi

  # shellcheck disable=SC2048
  # shellcheck disable=SC2086
  echo gcloud $* "${ZZ}" "${SA[@]}"

  # shellcheck disable=SC2236
  set +u
  if [[ -n "${DRY_RUN}" ]];then
    return
  fi
  set -u

  # shellcheck disable=SC2086
  # shellcheck disable=SC2048
  gcloud $* "${ZZ}" "${SA[@]}"
}

NETWORK_SUBNET="--create-subnetwork name=${CLUSTER_NAME}-subnet"
# shellcheck disable=SC2236
set +u
if [[ -n "${USE_SUBNET}" ]];then
  NETWORK_SUBNET="--network ${USE_SUBNET}"
fi
set -u

ADDONS="HorizontalPodAutoscaling,KubernetesDashboard"
# shellcheck disable=SC2236
set +u
if [[ -n "${ISTIO_ADDON}" ]];then
  ADDONS+=",Istio"
fi
set -u
# shellcheck disable=SC2086
# shellcheck disable=SC2046
gc beta container \
  --project "${PROJECT_ID}" \
  clusters create "${CLUSTER_NAME}" \
  --no-enable-basic-auth --cluster-version "${GKE_VERSION}" \
  --issue-client-certificate \
  --machine-type "${MACHINE_TYPE}" --image-type ${IMAGE} --disk-type "pd-standard" --disk-size "${DISK_SIZE}" \
  --scopes "${SCOPES}" \
  --num-nodes "${MIN_NODES}" --enable-autoscaling --min-nodes "${MIN_NODES}" --max-nodes "${MAX_NODES}" --max-pods-per-node "${MAXPODS_PER_NODE}" \
  --enable-stackdriver-kubernetes \
  --enable-ip-alias \
  --metadata disable-legacy-endpoints=true \
  ${NETWORK_SUBNET} \
  --default-max-pods-per-node "${MAXPODS_PER_NODE}" \
  --addons "${ADDONS}" \
  --enable-network-policy \
  --workload-metadata-from-node=EXPOSED \
  --enable-autoupgrade --enable-autorepair \
  --labels csm=1,test-date=$(date +%Y-%m-%d),version=${ISTIO_VERSION},operator=user_${USER}

NETWORK_NAME=$(basename "$(gcloud container clusters describe "${CLUSTER_NAME}" --project "${PROJECT_ID}" --zone="${ZONE}" \
    --format='value(networkConfig.network)')")
SUBNETWORK_NAME=$(basename "$(gcloud container clusters describe "${CLUSTER_NAME}" --project "${PROJECT_ID}" \
    --zone="${ZONE}" --format='value(networkConfig.subnetwork)')")

# Getting network tags is painful. Get the instance groups, map to an instance,
# and get the node tag from it (they should be the same across all nodes -- we don't
# know how to handle it, otherwise).
INSTANCE_GROUP=$(gcloud container clusters describe "${CLUSTER_NAME}" --project "${PROJECT_ID}" --zone="${ZONE}" --format='flattened(nodePools[].instanceGroupUrls[].scope().segment())' |  cut -d ':' -f2 | head -n1 | sed -e 's/^[[:space:]]*//' -e 's/::space:]]*$//')
INSTANCE_GROUP_ZONE=$(gcloud compute instance-groups list --filter="name=(${INSTANCE_GROUP})" --format="value(zone)" | sed 's|^.*/||g')
sleep 1
INSTANCE=$(gcloud compute instance-groups list-instances "${INSTANCE_GROUP}" --project "${PROJECT_ID}" \
    --zone="${INSTANCE_GROUP_ZONE}" --format="value(instance)" --limit 1)
NETWORK_TAGS=$(gcloud compute instances describe "${INSTANCE}" --zone="${INSTANCE_GROUP_ZONE}" --project "${PROJECT_ID}" --format="value(tags.items)")


NEGZONE=""
if [[ -n "${REGION}" ]]; then
  NEGZONE="region = ${REGION}"
else
  NEGZONE="local-zone = ${ZONE}"
fi

cat <<EOF > "${CLUSTER_NAME}/configmap-neg.yaml"
apiVersion: v1
kind: ConfigMap
metadata:
  name: gce-config
  namespace: kube-system
data:
  gce.conf: |
    [global]
    token-url = nil
    # Your cluster's project
    project-id = ${PROJECT_ID}
    # Your cluster's network
    network-name =  ${NETWORK_NAME}
    # Your cluster's subnetwork
    subnetwork-name = ${SUBNETWORK_NAME}
    # Prefix for your cluster's IG
    node-instance-prefix = gke-${CLUSTER_NAME}
    # Network tags for your cluster's IG
    node-tags = ${NETWORK_TAGS}
    # Zone the cluster lives in
    ${NEGZONE}
EOF


cat <<EOF > "${CLUSTER_NAME}/configmap-galley.yaml"
apiVersion: v1
kind: ConfigMap
metadata:
  name: istiod-asm
  namespace: istio-system
data:
  galley.json: |
      {
      "EnableServiceDiscovery": true,
      "SinkAddress": "meshconfig.googleapis.com:443",
      "SinkAuthMode": "GOOGLE",
      "ExcludedResourceKinds": ["Pod", "Node", "Endpoints"],
      "sds-path": "/etc/istio/proxy/SDS",
      "SinkMeta": ["project_id=${PROJECT_ID}"]
      }

  PROJECT_ID: ${PROJECT_ID}
  GOOGLE_APPLICATION_CREDENTIALS: /var/secrets/google/key.json
  ISTIOD_ADDR: istiod-asm.istio-system.svc:15012
  WEBHOOK: istiod-asm
  AUDIENCE: ${PROJECT_ID}.svc.id.goog
  trustDomain: ${PROJECT_ID}.svc.id.goog
  gkeClusterUrl: https://container.googleapis.com/v1/projects/${PROJECT_ID}/locations/${ZONE}/clusters/${CLUSTER_NAME}
EOF

export KUBECONFIG="${CLUSTER_NAME}/kube.yaml"
gcloud container clusters get-credentials "${CLUSTER_NAME}" --zone "${ZONE}"

kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole=cluster-admin \
  --user="$(gcloud config get-value core/account)"

# Update the cluster with the GCP-specific configmaps
kubectl -n kube-system apply -f "${CLUSTER_NAME}/configmap-neg.yaml"
kubectl -n kube-system create secret generic google-cloud-key  --from-file key.json="${CLUSTER_NAME}/google-cloud-key.json"

kubectl create ns istio-system
kubectl -n istio-system create secret generic google-cloud-key  --from-file key.json="${CLUSTER_NAME}/google-cloud-key.json"
kubectl -n istio-system apply -f "${CLUSTER_NAME}/configmap-galley.yaml"

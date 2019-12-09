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
# Creates service accounts to associate with a cluster with proper permissions.
#
# A control plane service account will also be created for interacting with GKE.
set -euo pipefail

# shellcheck disable=SC2086
WD=$(dirname $0)
# shellcheck disable=SC2086
WD=$(cd $WD || exit; pwd)

function gc() {
  # shellcheck disable=SC2048
  # shellcheck disable=SC2086
  echo gcloud $*

  # shellcheck disable=SC2236
  set +u
  if [[ -n "${DRY_RUN}" ]];then
    return
  fi
  set -u

  # shellcheck disable=SC2086
  # shellcheck disable=SC2048
  gcloud $*
}

PROJECT_ID=${PROJECT_ID:?"project id is required"}
GCP_SA=${1:-istio-data}
GCP_CTL_SA=${2:-istio-control}

gc iam service-accounts create "${GCP_SA}" --project "${PROJECT_ID}" # --display-name 'Istio data plane account'
gc iam service-accounts create "${GCP_CTL_SA}" --project "${PROJECT_ID}" #--display-name '"Istio control plane account"'

for role in compute.networkViewer logging.logWriter monitoring.metricWriter storage.objectViewer cloudtrace.agent meshtelemetry.reporter; do
	gc projects add-iam-policy-binding "${PROJECT_ID}" --role "roles/${role}" --member "serviceAccount:${GCP_SA}@${PROJECT_ID}.iam.gserviceaccount.com"
done

gc projects add-iam-policy-binding "${PROJECT_ID}" --role "roles/meshconfig.writer" --member "serviceAccount:${GCP_CTL_SA}@${PROJECT_ID}.iam.gserviceaccount.com"
# Required for creating the NEG objects in GCP
gc projects add-iam-policy-binding "${PROJECT_ID}" --role "roles/compute.admin" --member "serviceAccount:${GCP_CTL_SA}@${PROJECT_ID}.iam.gserviceaccount.com"

if [[ "${CLUSTER_NAME}" != "" ]]; then 
  gc iam service-accounts keys create "${WD}/tmp/${CLUSTER_NAME}/google-cloud-key.json" --iam-account="${GCP_CTL_SA}"@"${PROJECT_ID}".iam.gserviceaccount.com
fi

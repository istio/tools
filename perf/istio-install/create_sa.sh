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

if [[ -z "$(gcloud beta iam service-accounts --project="${PROJECT_ID}" list --filter=email="${GCP_SA}@${PROJECT_ID}.iam.gserviceaccount.com" --format='csv[no-heading](email)')" ]]; then
  gc beta iam service-accounts create --quiet "${GCP_SA}" # --display-name 'Istio data plane account'
fi
if [[ -z "$(gcloud beta iam service-accounts --project="${PROJECT_ID}" list --filter=email="${GCP_CTL_SA}@${PROJECT_ID}.iam.gserviceaccount.com" --format='csv[no-heading](email)')" ]]; then
  gc beta iam service-accounts create --quiet "${GCP_CTL_SA}" #--display-name '"Istio control plane account"'
fi

function gcloud_get_iam {
  local sa="${1?"Must pass a SA as first argument to gcloud_get_iam function"}"
  local role="${2?"Must pass a role as second argument to gcloud_get_iam function"}"
  gcloud projects get-iam-policy ${PROJECT_ID} --format=json | jq ".bindings[] | select((.members[] | contains(\"serviceAccount:${sa}@${PROJECT_ID}.iam.gserviceaccount.com\")) and .role == \"roles/${role}\")| .members[0]" -Mr

}

for role in compute.networkViewer logging.logWriter monitoring.metricWriter storage.objectViewer cloudtrace.agent meshtelemetry.reporter; do
  if [[ -z "$(gcloud_get_iam ${GCP_SA} ${role})" ]]; then 
	  gc projects add-iam-policy-binding --quiet "${PROJECT_ID}" --role "roles/${role}" --member "serviceAccount:${GCP_SA}@${PROJECT_ID}.iam.gserviceaccount.com"
  fi
done

if [[ -z "$(gcloud_get_iam ${GCP_CTL_SA} meshconfig.writer )" ]]; then
  gc projects add-iam-policy-binding --quiet "${PROJECT_ID}" --role "roles/meshconfig.writer" --member "serviceAccount:${GCP_CTL_SA}@${PROJECT_ID}.iam.gserviceaccount.com"
fi
# Required for creating the NEG objects in GCP
if [[ -z "$(gcloud_get_iam ${GCP_CTL_SA} compute.admin )" ]]; then
  gc projects add-iam-policy-binding --quiet "${PROJECT_ID}" --role "roles/compute.admin" --member "serviceAccount:${GCP_CTL_SA}@${PROJECT_ID}.iam.gserviceaccount.com"
fi


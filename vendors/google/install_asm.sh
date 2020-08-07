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

set -ex

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")

# Install ASM with kpt and istioctl on an existing gke cluster
# This may not work with older version of ASM before 1.6.
# You might need to Set ASM_HUB and ASM_TAG env variables accordingly, otherwise default values would be used.
function install_asm() {
  if [[ -z "${PROJECT_ID}" ]] || [[ -z "${CLUSTER_NAME}" ]] || [[ -z "${CLUSTER_LOCATION}" ]] || [[ -z "${RELEASE}" ]]  ;then
    echo "You need to set these env variables first: PROJECT_ID, CLUSTER_NAME, CLUSTER_LOCATION, RELEASE"
    echo "for example PROJECT_ID=test_proj CLUSTER_NAME=test CLUSTER_LOCATION=us-central1-a RELEASE=release-1.6-asm"
    exit 1
  fi

  TMP_DIR=$(mktemp -d -t "${RELEASE}-XXXXXX")
  trap 'rm -rf "${TMP_DIR}"' EXIT
  cd "${TMP_DIR}"

  export PROJECT_NUMBER=$(gcloud projects describe ${PROJECT_ID} --format="value(projectNumber)")
  export WORKLOAD_POOL=${PROJECT_ID}.svc.id.goog
  export MESH_ID="proj-${PROJECT_NUMBER}"
  export ASM_HUB="${ASM_HUB:-}"
  export ASM_TAG="${ASM_TAG:-}"

  gcloud config set compute/zone ${CLUSTER_LOCATION}
  gcloud container clusters update ${CLUSTER_NAME} --update-labels=mesh_id=${MESH_ID}
  gcloud container clusters update ${CLUSTER_NAME} --workload-pool=${WORKLOAD_POOL}
  gcloud container clusters update ${CLUSTER_NAME} --enable-stackdriver-kubernetes

  curl --request POST \
  --header "Authorization: Bearer $(gcloud auth print-access-token)" \
  --data '' \
  "https://meshconfig.googleapis.com/v1alpha1/projects/${PROJECT_ID}:initialize"
  gcloud container clusters get-credentials ${CLUSTER_NAME}
  kubectl create clusterrolebinding cluster-admin-binding \
  --clusterrole=cluster-admin \
  --user="$(gcloud config get-value core/account)" || true

  kpt pkg get "https://github.com/GoogleCloudPlatform/anthos-service-mesh-packages.git/asm@${RELEASE}" .
  if [[ -n "${ASM_HUB}" ]];then
    kpt cfg set asm anthos.servicemesh.tag ${ASM_HUB}
  fi
  if [[ -n "${ASM_TAG}" ]];then
    kpt cfg set asm anthos.servicemesh.hub ${ASM_TAG}
  fi
  kpt cfg set asm gcloud.container.cluster ${CLUSTER_NAME}
  kpt cfg set asm gcloud.core.project ${PROJECT_ID}
  kpt cfg set asm gcloud.compute.location ${CLUSTER_LOCATION}

  "${release}/bin/istioctl" install -f asm/cluster/istio-operator.yaml -f "${WD}/istioctl_profiles/long-running-asm.yaml" "${@}"
  install_extras
}

# ASM installation
export INSTALL_ASM="${INSTALL_ASM:-}"
export MULTI_CLUSTER="${MULTI_CLUSTER:-}"
export SKIP_INSTALL=true
source ${ROOT}/../perf/istio-install/setup_istio.sh

download_release
release="${DIRNAME}/${OUT_FILE}"
if [[ -n "${MULTI_CLUSTER}" ]];then
  kubectl config use-context "${CTX1}"
  CLUSTER_NAME="${CLUSTER1}" install_asm "${@}"
  kubectl config use-context "${CTX2}"
  CLUSTER_NAME="${CLUSTER2}" install_asm "${@}"
  "${release}/bin/istioctl" x create-remote-secret --context=${CTX1} --name=${CLUSTER1} | \
  kubectl apply -f - --context=${CTX2}
  "${release}/bin/istioctl" x create-remote-secret --context=${CTX2} --name=${CLUSTER2} | \
  kubectl apply -f - --context=${CTX1}
else
  install_asm "${@}"
fi
exit 0

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
DIRNAME="${WD}/tmp"
mkdir -p "${DIRNAME}"
export GO111MODULE=on

case "${OSTYPE}" in
  darwin*) export ARCH_SUFFIX="${ARCH_SUFFIX:-osx}" ;;
  linux*) export ARCH_SUFFIX="${ARCH_SUFFIX:-linux-amd64}" ;;
  *) echo "unsupported: ${OSTYPE}" ;;
esac

# We support many ways to reference the source to install from
# This logic ultimately sets up the following variables:
# * OUT_FILE: where the download will be stored. Note this is cached.
# * RELEASE_URL: where to download the release from.

# Passing a tag, like latest or 1.4-dev
if [[ -n "${TAG:-}" ]]; then
  VERSION=$(curl -sL https://gcsweb.istio.io/gcs/istio-build/dev/"${TAG}")
  OUT_FILE="istio-${VERSION}"
  RELEASE_URL="https://storage.googleapis.com/istio-build/dev/${VERSION}/istio-${VERSION}-${ARCH_SUFFIX}.tar.gz"
# Passing a dev version, like 1.4-alpha.41dee99277dbed4bfb3174dd0448ea941cf117fd
elif [[ -n "${DEV_VERSION:-}" ]]; then
  OUT_FILE="istio-${DEV_VERSION}"
  RELEASE_URL="https://storage.googleapis.com/istio-build/dev/${DEV_VERSION}/istio-${DEV_VERSION}-${ARCH_SUFFIX}.tar.gz"
# Passing a version, like 1.4.2
elif [[ -n "${VERSION:-}" ]]; then
  OUT_FILE="istio-${VERSION}"
  RELEASE_URL="https://storage.googleapis.com/istio-prerelease/prerelease/${VERSION}/istio-${VERSION}-${ARCH_SUFFIX}.tar.gz"
# Passing a release url, like https://storage.googleapis.com/istio-prerelease/prerelease/1.4.1/istio-1.4.1-linux-amd64.tar.gz
elif [[ -n "${RELEASE_URL:-}" ]]; then
  OUT_FILE=${OUT_FILE:-"$(basename "${RELEASE_URL}" "-${ARCH_SUFFIX}.tar.gz")"}
# Passing a gcs url, like gs://istio-build/dev/1.4-alpha.41dee99277dbed4bfb3174dd0448ea941cf117fd
elif [[ -n "${GCS_URL:-}" ]]; then
  RELEASE_URL="${GCS_URL}"
  OUT_FILE=${OUT_FILE:-"$(basename "${RELEASE_URL}" "-${ARCH_SUFFIX}.tar.gz")"}
fi

if [[ -z "${RELEASE_URL:-}" ]]; then
  echo "Must set one of TAG, VERSION, DEV_VERSION, RELEASE_URL, GCS_URL"
  exit 2
fi

function download_release() {
  outfile="${DIRNAME}/${OUT_FILE}"
  if [[ ! -d "${outfile}" ]]; then
    tmp=$(mktemp -d)
    if [[ "${RELEASE_URL}" == gs://* ]]; then
      gsutil cp "${RELEASE_URL}" "${tmp}/out.tar.gz"
      tar xvf "${tmp}/out.tar.gz" -C "${DIRNAME}"
    else
      curl -fJLs -o "${tmp}/out.tar.gz" "${RELEASE_URL}"
      tar xvf "${tmp}/out.tar.gz" -C "${DIRNAME}"
    fi
  else
    echo "${outfile} already exists, skipping download"
  fi
}

function install_istioctl() {
  release=${1:?release folder}
  shift
  "${release}/bin/istioctl" install --skip-confirmation -d "${release}/manifests" "${@}"
}

function install_extras() {
  local domain=${DNS_DOMAIN:-"DNS_DOMAIN like v104.qualistio.org"}
  kubectl create namespace istio-prometheus || true
  # Deploy the gateways and prometheus operator.
  # We install the prometheus operator first, then deploy the CR, to wait for the CRDs to get created
  helm template --set domain="${domain}" --set prometheus.deploy=false "${WD}/base" | kubectl apply -f -
  # Check CRD
  CMDs_ARR=('kubectl get crds/prometheuses.monitoring.coreos.com' 'kubectl get crds/alertmanagers.monitoring.coreos.com'
  'kubectl get crds/podmonitors.monitoring.coreos.com' 'kubectl get crds/prometheusrules.monitoring.coreos.com'
  'kubectl get crds/servicemonitors.monitoring.coreos.com')
  for CMD in "${CMDs_ARR[@]}"
  do
    MAXRETRIES=0
    until $CMD || [ $MAXRETRIES -eq 60 ]
    do
      MAXRETRIES=$((MAXRETRIES + 1))
      sleep 5
    done
    if [[ $MAXRETRIES -eq 60 ]]; then
      echo "crds were not created successfully"
      exit 1
    fi
  done
  # Redeploy, this time with the Prometheus resource created
  helm template --set domain="${domain}" "${WD}/base" | kubectl apply -f -
  # Also deploy relevant ServiceMonitors
  if [[ -f "${release}/samples/addons/extras/prometheus-operator.yaml" ]];then
     kubectl apply -f "${release}/samples/addons/extras/prometheus-operator.yaml" -n istio-system
  # for release before 1.7, run below instead
  else
    "${release}/bin/istioctl" manifest generate --set profile=empty --set addonComponents.prometheusOperator.enabled=true -d "${release}/manifests" | kubectl apply -f -
  fi
  # deploy grafana
  kubectl apply -f "${release}/samples/addons/grafana.yaml" -n istio-system
}

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
if [[ -n "${INSTALL_ASM}" ]];then
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
fi

if [[ -z "${LOCAL_ISTIO_PATH}" ]];then
  download_release
  install_istioctl "${DIRNAME}/${OUT_FILE}" "${@}"

  if [[ -z "${SKIP_EXTRAS:-}" ]]; then
    install_extras
  fi
# if LOCAL_ISTIO_PATH is set, we assume that Istio is preconfigured, we only install extra monitoring/alerting configs.
else
  release="${LOCAL_ISTIO_PATH}"
  install_extras
fi

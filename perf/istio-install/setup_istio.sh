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

IOPS="${IOPS:-istioctl_profiles/long-running.yaml,istioctl_profiles/long-running-gateway.yaml}"

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
  for i in ${IOPS//,/ }; do
    "${release}/bin/istioctl" install --skip-confirmation -d "${release}/manifests" -f "${i}" "${@}"
  done
}

function install_extras() {
  local domain=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org"}
  local certmanagerEmail=${CERTMANAGER_EMAIL:-""}
  local helmValues=${HELM_VALUES:-}
  kubectl create namespace istio-prometheus || true
  # Deploy the gateways and prometheus operator.
  # Deploy CRDs with create, they are too big otherwise
  kubectl create -f base/files || true # Might fail if we already installed, so allow failures
  if [[ "${certmanagerEmail:-}" != "" ]]; then
    kubectl apply -f "${WD}/addons/cert-manager.yaml"
    kubectl wait --for=condition=Available deployments --all -n cert-manager
    helm template --set domain="${domain}" --set certManager.email="${certmanagerEmail}" --set certManager.enabled=true "${WD}/base" | kubectl apply -f -
  else
    # shellcheck disable=SC2086
    helm template --set domain="${domain}" ${helmValues} "${WD}/base" | kubectl apply -f -
  fi

  # Check deployment
  MAXRETRIES=0
  until kubectl rollout status --watch --timeout=60s statefulset/prometheus-prometheus -n istio-prometheus || [ $MAXRETRIES -eq 60 ]
  do
    MAXRETRIES=$((MAXRETRIES + 1))
    sleep 5
  done
  if [[ $MAXRETRIES -eq 60 ]]; then
    echo "prometheus were not created successfully"
    exit 1
  fi

  # Also deploy relevant ServiceMonitors
  if [[ -f "${release}/samples/addons/extras/prometheus-operator.yaml" ]];then
     kubectl apply -f "${release}/samples/addons/extras/prometheus-operator.yaml" -n istio-system
      # Deploy k8s ServiceMonitors
     kubectl apply -f "${WD}/addons/servicemonitors.yaml"
  # for release before 1.7, run below instead
  else
    "${release}/bin/istioctl" manifest generate --set profile=empty --set addonComponents.prometheusOperator.enabled=true -d "${release}/manifests" | kubectl apply -f -
  fi
  # deploy grafana
  kubectl apply -f "${release}/samples/addons/grafana.yaml" -n istio-system
  kubectl apply -f "${WD}/addons/grafana-cm.yaml" -n istio-system # override just the configmap
  kubectl rollout restart deployment grafana -n istio-system # restart to ensure it picks up our new configmap
}

if [[ -z "${SKIP_INSTALL}" ]];then
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
fi

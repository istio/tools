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

DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org"}

# shellcheck disable=SC2086
WD=$(dirname $0)
# shellcheck disable=SC2086
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"
export INSTALL_WITH_ISTIOCTL=true

release="${1:?"release"}"

if [[ "${release}" == *-latest ]];then
  # shellcheck disable=SC2086
  release=$(curl -f -L https://storage.googleapis.com/istio-prerelease/daily-build/${release}.txt)
  # shellcheck disable=SC2181
  if [[ $? -ne 0 ]];then
    echo "${release} branch does not exist"
    exit 1
  fi
fi

shift

function download() {
  local DIRNAME="$1"
  local release="$2"

  local url="https://gcsweb.istio.io/gcs/istio-build/dev/${release}/istio-${release}-linux.tar.gz"
  # shellcheck disable=SC2236
  if [[ -n "${RELEASE_URL}" ]];then
    url="${RELEASE_URL}"
  fi
  local outfile="${DIRNAME}/istio-${release}.tgz"
  if [[ ! -f "${outfile}" ]]; then
    # shellcheck disable=SC2091
    # shellcheck disable=SC2086
    $(curl -fJL -o "${outfile}" ${url})
    # shellcheck disable=SC2181
    if [[ $? -ne 0 ]];then
      echo ""
      exit 1
    fi
  fi

  echo "${outfile}"
}

function trim(){
    if [[ "$1" =~ [^[:space:]](.*[^[:space:]])? ]]; then
      echo "${BASH_REMATCH[0]}"
    fi
}

function setup_admin_binding() {
  # shellcheck disable=SC2046
  kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account) || true
}

# install istio with default IstioControlPlane CR yaml using istioctl.
function install_istio_with_istioctl() {
  local CR_PATH="${WD}/istioctl_profiles/${CR_FILENAME}"
  pushd "${ISTIOCTL_PATH}"
  # shellcheck disable=SC2086
  ./istioctl manifest apply -f "${CR_PATH}" --set "${SET_OVERLAY}" ${EXTRA_ARGS} --wait
  popd
}

function install_istio() {
  local DIRNAME="${1:?"output dir"}"
  local release="${2:?"release"}"
  # shellcheck disable=SC2155
  local release_ver=$(echo "$release" | cut -f1 -d "-")
  # shellcheck disable=SC2072
  if [[ "${release_ver}" > "1.5" ]];then
    export INSTALL_WITH_ISTIOCTL=true
  fi
  # shellcheck disable=SC2155
  # shellcheck disable=SC2086
  local outfile="$(download ${DIRNAME} ${release})"
  if [[ "$outfile" == "" ]];then
    echo "failed to download istio release"
    exit 1
  fi
  # shellcheck disable=SC2086
  outfile=$(trim $outfile)

  if [[ ! -d "${DIRNAME}/${release}" ]];then
      DN=$(mktemp -d)
      tar -xzf "${outfile}" -C "${DN}" --strip-components 1
      if [[ -z "${INSTALL_WITH_ISTIOCTL}" ]]; then
        mv "${DN}/install/kubernetes/helm" "${DIRNAME}/${release}"
        mv "${DN}/bin/istioctl" "${DIRNAME}/${release}"
      fi
      cp "${DN}/bin/istioctl" "${DIRNAME}"
      rm -rf "${DN}"
  fi

  export ISTIOCTL_PATH="${DIRNAME}"

  if [[ -n "${SKIP_INSTALLATION}" ]];then
    echo "skip installation step"
    return
  fi
  if [[ -z "${INSTALL_WITH_ISTIOCTL}" ]]; then
    echo "start installing istio using helm"
    install_istio_with_helm
  else
    echo "start installing istio using istioctl"
    export SET_OVERLAY="meshConfig.rootNamespace=istio-system"
    if [[ -z "${CR_FILENAME}" ]]; then
      CR_FILENAME="default.yaml"
    fi
    install_istio_with_istioctl
  fi
}

function install_gateways() {
  local domain=${DNS_DOMAIN:-qualistio.org}
  if [[ -z "${DRY_RUN}" ]]; then
      helm template --set domain="${domain}" base | kubectl -n istio-system apply -f -
  fi
}

setup_admin_binding
# shellcheck disable=SC2048
# shellcheck disable=SC2086
install_istio "${WD}/tmp" "${release}" $*
install_gateways

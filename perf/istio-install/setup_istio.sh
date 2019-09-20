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

  local url="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${release}/istio-${release}-linux.tar.gz"
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

function install_istio() {
  local DIRNAME="${1:?"output dir"}"
  local release="${2:?"release"}"

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
      mv "${DN}/install/kubernetes/helm" "${DIRNAME}/${release}"
      # shellcheck disable=SC2086
      rm -rf ${DN}
  fi

  kubectl create ns istio-system || true

  if [[ -z "${DRY_RUN}" ]];then
      # apply CRD files for istio kinds
      if [[ -f "${DIRNAME}/${release}/istio/templates/crds.yaml" ]];then
         kubectl apply -f "${DIRNAME}/${release}/istio/templates/crds.yaml"
         kubectl wait --for=condition=Established -f "${DIRNAME}/${release}/istio/templates/crds.yaml"
      else
         kubectl apply -f "${DIRNAME}/${release}/istio-init/files/"
         kubectl wait --for=condition=Established -f "${DIRNAME}/${release}/istio-init/files/"
      fi
  fi

  local FILENAME="${DIRNAME}/${release}.yml"

  # if release_url is not overridden then daily builds require
  # tag and hub overrides

  if [[ -z "${RELEASE_URL}" ]];then
    opts+=" --set global.tag=${release}"
    opts+=" --set global.hub=gcr.io/istio-release"
  fi

  local values=${VALUES:-values.yaml}
  local extravalues=${EXTRA_VALUES:-""}
  if [[ ${extravalues} != "" ]]; then
    extravalues="--values ${extravalues}"
  fi

  # shellcheck disable=SC2086
  helm template --name istio --namespace istio-system \
       ${opts} \
       --values ${values} \
       ${extravalues} \
       "${DIRNAME}/${release}/istio" > "${FILENAME}"

  if [[ -z "${DRY_RUN}" ]];then
      kubectl apply -f "${FILENAME}"
      if [[ -z "${SKIP_PROMETHEUS}" ]];then
          # shellcheck disable=SC2086
          "$WD/setup_prometheus.sh" ${DIRNAME}
      fi
  fi

  echo "Wrote file ${FILENAME}"
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

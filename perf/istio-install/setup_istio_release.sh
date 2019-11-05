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
# shellcheck disable=SC2086
mkdir -p "${WD}/tmp"

release="${1:?"release"}"
stream=""

shift

if [[ "${release}" == *-dev ]];then
  # shellcheck disable=SC2086
  release=$(curl -f -L https://storage.googleapis.com/istio-build/dev/${release})
  # shellcheck disable=SC2181
  if [[ $? -ne 0 ]];then
    echo "${release} does not exist"
    exit 1
  fi
  stream="dev"
else
  stream="$1"
fi


if [[ "${stream}" == "pre-release" ]];then
  export HELMREPO_URL=https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/charts
  case "${OSTYPE}" in
    darwin*) export RELEASE_URL=https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/istio-${release}-osx.tar.gz ;;
    linux*) export RELEASE_URL=https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/istio-${release}-linux.tar.gz ;;
    *) echo "unsupported: ${OSTYPE}" ;;
  esac
elif [[ "${stream}" == "dev" ]];then
  export HELMREPO_URL=https://gcsweb.istio.io/gcs/istio-build/dev/${release}/charts
  case "${OSTYPE}" in
    darwin*) export RELEASE_URL=https://gcsweb.istio.io/gcs/istio-build/dev/${release}/istio-${release}-osx.tar.gz ;;
    linux*) export RELEASE_URL=https://gcsweb.istio.io/gcs/istio-build/dev/${release}/istio-${release}-linux.tar.gz ;;
    *) echo "unsupported: ${OSTYPE}" ;;
  esac
else
  export HELMREPO_URL=https://storage.googleapis.com/istio-release/releases/${release}/charts
  case "${OSTYPE}" in
    darwin*) export RELEASE_URL=https://github.com/istio/istio/releases/download/${release}/istio-${release}-osx.tar.gz ;;
    linux*) export RELEASE_URL=https://github.com/istio/istio/releases/download/${release}/istio-${release}-linux.tar.gz ;;
    *) echo "unsupported: ${OSTYPE}" ;;
  esac
fi

# shellcheck disable=SC2086
${WD}/setup_istio.sh "${release}"

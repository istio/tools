#!/bin/bash
set -ex

DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org"}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

release="${1:?"release"}"
shift

if [[ "$1" == "pre-release" ]];then
  export HELMREPO_URL=https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/charts
  case "${OSTYPE}" in
    darwin*) export RELEASE_URL=https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/istio-${release}-osx.tar.gz ;;
    linux*) export RELEASE_URL=https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/istio-${release}-linux.tar.gz ;;
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

${WD}/setup_istio.sh "${release}"
#!/bin/bash
set -ex

DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org"}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

release="${1:?"release"}"
shift

export RELEASE_URL=https://github.com/istio/istio/releases/download/${release}/istio-${release}-linux.tar.gz
export HELMREPO_URL=https://storage.googleapis.com/istio-release/releases/${release}/charts

${WD}/setup_istio.sh "${release}"

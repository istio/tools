#!/bin/bash
set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

git clone https://github.com/istio/tools
cd tools/perf/istio-install
helm init --client-only
TARGET_VERSION=$1
LATEST_BUILD=$(curl -L https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/release-$TARGET_VERSION-latest.txt)
HELMREPO_URL=https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/$LATEST_BUILD/charts/ RELEASE_URL=https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/$LATEST_BUILD/istio-$LATEST_BUILD-linux.tar.gz DNS_DOMAIN=v11.qualistio.org ./setup_istio.sh $TARGET_VERSION

# trigger redeploy on services to get new sidecars
${WD}/../../bin/redeploy.sh ALL

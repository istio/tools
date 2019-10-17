#!/bin/bash
# 
# Creates a asan-build image based on the published proxyv2 image.
# 
# Uses ASAN build from PROXY_REPO_SHA embedded in proxyv2 image unless a different PROXY_REPO_SHA is specified.
#
# Example: ./package_asan_image.sh 1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e

set  -ex

WD=$(dirname $0)
WD=$(cd ${WD}; pwd)

ISTIO_SHA=${1:?"ISTIO SHA for proxyv2 required"}
HUB=${HUB:-"gcr.io/mixologist-142215"}

BASE_IMAGE="gcr.io/istio-testing/proxyv2:${ISTIO_SHA}"
if [[ -z "${PROXY_REPO_SHA}" ]];then
  docker pull "${BASE_IMAGE}"
  PROXY_REPO_SHA=$(docker image inspect "${BASE_IMAGE}" --format="{{range .ContainerConfig.Env}}{{println .}}{{end}}" | grep ISTIO_META_ISTIO_PROXY_SHA | cut -d ":" -f 2)
fi

cd "${WD}"
mkdir -p "${ISTIO_SHA}"

PROXY_TGZ="${ISTIO_SHA}/envoy.tgz"
ASAN_URL="https://storage.googleapis.com/istio-build/proxy/envoy-asan-${PROXY_REPO_SHA}.tar.gz"

rm -f "${PROXY_TGZ}"
curl -fJL -o "${PROXY_TGZ}" "${ASAN_URL}"

cd "${WD}/${ISTIO_SHA}"
ASAN_IMAGE="${HUB}/proxyv2:asan-${ISTIO_SHA}"

docker build -f ../Dockerfile --build-arg "BASE_IMAGE=${BASE_IMAGE}" --build-arg "proxy_version=istio-proxy:${PROXY_REPO_SHA}" -t "${ASAN_IMAGE}" .
docker push "${ASAN_IMAGE}"

# Sample run
# 
#++ dirname ./package_asan_image.sh
#+ WD=.
#++ cd .
#++ pwd
#+ WD=/home/mjog/newbuild
#+ ISTIO_SHA=1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e
#+ HUB=gcr.io/mixologist-142215
#+ BASE_IMAGE=gcr.io/istio-testing/proxyv2:1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e
#+ [[ -z '' ]]
#+ docker pull gcr.io/istio-testing/proxyv2:1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e
#1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e: Pulling from istio-testing/proxyv2
#Digest: sha256:353bee2bd7dfaf25cfe4439e9befd25f00f1c56fcbf7bc6cae43d8791430d99b
#Status: Downloaded newer image for gcr.io/istio-testing/proxyv2:1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e
#++ docker image inspect gcr.io/istio-testing/proxyv2:1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e '--format={{range .ContainerConfig.Env}}{{println .}}{{end}}'
#++ grep ISTIO_META_ISTIO_PROXY_SHA
#++ cut -d : -f 2
#
#+ PROXY_REPO_SHA=08cec447dfc349b314e226dc66132dbdb5d08fc9
#+ cd /home/mjog/newbuild
#+ mkdir -p 1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e
#+ PROXY_TGZ=1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e/envoy.tgz
#+ ASAN_URL=https://storage.googleapis.com/istio-build/proxy/envoy-asan-08cec447dfc349b314e226dc66132dbdb5d08fc9.tar.gz
#+ rm -f 1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e/envoy.tgz
#+ curl -fJL -o 1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e/envoy.tgz https://storage.googleapis.com/istio-build/proxy/envoy-asan-08cec447dfc349b314e226dc66132dbdb5d08fc9.t
#ar.gz
#  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
#                                 Dload  Upload   Total   Spent    Left  Speed
#100  205M  100  205M    0     0   165M      0  0:00:01  0:00:01 --:--:--  165M
#+ cd /home/mjog/newbuild/1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e
#+ ASAN_IMAGE=gcr.io/mixologist-142215/proxyv2:asan-1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e
#+ docker build -f ../Dockerfile --build-arg BASE_IMAGE=gcr.io/istio-testing/proxyv2:1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e --build-arg proxy_version=istio-proxy:08cec
#447dfc349b314e226dc66132dbdb5d08fc9 -t gcr.io/mixologist-142215/proxyv2:asan-1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e .
#Sending build context to Docker daemon  215.7MB
#Step 1/8 : ARG BASE_IMAGE
#Step 2/8 : ARG proxy_version
#Step 3/8 : FROM ${BASE_IMAGE}
# ---> 8d37f57f1c44
#Step 4/8 : ENV ISTIO_META_ISTIO_PROXY_SHA $proxy_version
# ---> Running in cf077aa28681
#Removing intermediate container cf077aa28681
# ---> 28ae1132c522
#Step 5/8 : RUN apt-get update && apt-get install -y llvm libc++-dev
#Setting up llvm-6.0-dev (1:6.0-1ubuntu2) ...
#Processing triggers for libc-bin (2.27-3ubuntu1) ...
#Removing intermediate container 31c313174afe
# ---> 4a70a0de4a85
#Step 6/8 : ENV ASAN_SYMBOLIZER_PATH /usr/bin/llvm-symbolizer
# ---> Running in a2ada88f1844
#Removing intermediate container a2ada88f1844
# ---> c81bc6fe54fb
#Step 7/8 : ENV ASAN_OPTIONS halt_on_error=0
# ---> Running in 0f4925edd86b
#Removing intermediate container 0f4925edd86b
# ---> 10bec9ff6c00
#Step 8/8 : ADD envoy.tgz /
# ---> 63a2fa634e2f
#Successfully built 63a2fa634e2f
#Successfully tagged gcr.io/mixologist-142215/proxyv2:asan-1.5-alpha.565947c1dccb2fab7704ef1727212eabe2c9403e

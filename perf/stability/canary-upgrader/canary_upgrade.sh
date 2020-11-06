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

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
DIRNAME="/tmp"
set -eux

function download_release() {
# shellcheck disable=SC2155
  export VERSION=$(curl -sL https://gcsweb.istio.io/gcs/istio-build/dev/latest)
  OUT_FILE="istio-${VERSION}"
  RELEASE_URL="https://storage.googleapis.com/istio-build/dev/${VERSION}/istio-${VERSION}-linux-amd64.tar.gz"
  outfile="${DIRNAME}/${OUT_FILE}"
  if [[ ! -d "${outfile}" ]]; then
    tmp=$(mktemp -d)
    curl -fJLs -o "${tmp}/out.tar.gz" "${RELEASE_URL}"
    tar xvf "${tmp}/out.tar.gz" -C "${DIRNAME}"
  else
    echo "${outfile} already exists, skipping download"
  fi
}

function install_istioctl() {
  "${outfile}/bin/istioctl" install -d "${outfile}/manifests" --set revision="${NEW_REV}" --skip-confirmation
}

# existing revision
REV_LIST=$(kubectl get pods -n istio-system -lapp=istiod --sort-by=.status.startTime -o "jsonpath={.items[*].metadata.labels.istio\.io\/rev}")
EXISTING_REV=$(echo "${REV_LIST}" | cut -f1 -d' ')

download_release
SUFFIX=$(echo "${VERSION}" | cut -f2 -d- | cut -f2 -d.)
NEW_REV="canary-${SUFFIX}"
install_istioctl

# verify whether canary one exist
podc=$(kubectl -n istio-system get pods -l istio.io/rev="${NEW_REV}" | grep -c istiod)
svcc=$(kubectl -n istio-system get svc -l istio.io/rev="${NEW_REV}" | grep -c istiod)
if [[ ${podc} == 0 ]] || [[ ${svcc} == 0 ]]; then
  echo "canary deployment not available"
  exit 1
fi
allns=$(kubectl get ns -o jsonpath="{.items[*].metadata.name}")
# upgrade data plane
for testns in ${allns};do
    if [[ ${testns} == *"service-graph"* ]];then
        kubectl label namespace "${testns}" istio-injection- istio.io/rev="${NEW_REV}" --overwrite || true
        kubectl rollout restart deployment -n "${testns}"
        sleep 30
    # verify
    fi
done

# clean up old control plane
# This command only works for 1.7 or later
if [[ -n ${EXISTING_REV} ]];then
  "${outfile}/bin/istioctl" x uninstall --revision "${EXISTING_REV}"
fi

# update sha for the spanner table
# for ingress, /var/lib/istio/data is valid only for 1.8
kubectl set env deployment/am-webhook -n istio-prometheus BRANCH="${VERSION}"

# get memory profile
# shellcheck disable=SC2155
export POD=$(kubectl get pod -l app=istio-ingressgateway -o jsonpath="{.items[0].metadata.name}" -n istio-system)
export NS=istio-system
kubectl exec "${POD}" -n "${NS}" -c istio-proxy -- curl -X POST -s "http://localhost:15000/heapprofiler?enable=y"
sleep 15
kubectl exec "${POD}" -n "${NS}" -c istio-proxy -- curl -X POST -s "http://localhost:15000/heapprofiler?enable=n"

sleep 15
# get cpu profile
kubectl -n ${NS} exec "${POD}" -c istio-proxy -- sh -c 'sudo mkdir -p /var/log/envoy && sudo chmod 777 /var/log/envoy && curl -X POST -s "http://localhost:15000/cpuprofiler?enable=y"'
sleep 15
kubectl -n ${NS} exec "${POD}" -c istio-proxy -- sh -c 'curl -X POST -s "http://localhost:15000/cpuprofiler?enable=n"'
kubectl -n ${NS} cp "${POD}":/var/lib/istio/data /tmp/envoy -c istio-proxy
kubectl -n ${NS} cp "${POD}":/lib/x86_64-linux-gnu /tmp/envoy/lib -c istio-proxy
kubectl -n ${NS} cp "${POD}":/usr/local/bin/envoy /tmp/envoy/lib/envoy -c istio-proxy

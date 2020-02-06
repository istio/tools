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

set -e

function usage() {
  echo "usage:
        ./get_proxy_perf.sh -p <pod_name> -n <pod_namespace> -s <sample_frequency> -t <time>
    
    e.g.
      ./get_proxy_perf.sh -p svc05-0-4-0-67bff5dbbf-grl94 -n service-graph05 -s 177 -t 20

    -p name of pod.
    -n namespace of the given pod.
    -s sample frequence in Hz.
    -t time of profiling in second."
  exit 1
}

POD_NAME=""
POD_NAMESPACE=""
SAMPLE_FREQUENCY="177"
PERF_TIME="20"

while getopts p:n:s:t: arg ; do
  case "${arg}" in
    p) POD_NAME="${OPTARG}";;
    n) POD_NAMESPACE="${OPTARG}";;
    t) SAMPLE_FREQUENCY="${OPTARG}";;
    s) PERF_TIME="${OPTARG}";;
    *) usage;;
  esac
done

if [ -z "${POD_NAME}" ]; then
    echo "pod name must be provided."
    usage
    exit 1
fi

if [ -z "${POD_NAMESPACE}" ]; then
    echo "pod namespace must be provided."
    usage
    exit 1
fi

WD=$(dirname "${0}")
WD=$(cd "${WD}" && pwd)

echo "copy profiling script to proxy..."
kubectl cp "${WD}"/get_perfdata.sh "${POD_NAME}":/etc/istio/proxy/get_perfdata.sh -n "${POD_NAMESPACE}" -c istio-proxy

echo "start profiling..."
kubectl exec "${POD_NAME}" -n "${POD_NAMESPACE}" -c istio-proxy -- /etc/istio/proxy/get_perfdata.sh perf.data "${SAMPLE_FREQUENCY}" "${PERF_TIME}"

TMP_DIR=$(mktemp -d -t proxy-perf-XXXXXXXXXX)
trap 'rm -rf "${TMP_DIR}"' EXIT
TIME="$(date '+%Y-%m-%d-%H-%M-%S')"
PERF_FILE_NAME="${POD_NAME}"-"${TIME}".perf
PERF_FILE="${TMP_DIR}"/"${PERF_FILE_NAME}"
kubectl cp "${POD_NAME}":/etc/istio/proxy/perf.data.perf "${PERF_FILE}" -n "${POD_NAMESPACE}" -c istio-proxy

echo "generating svg file ${PERF_FILE_NAME}"
"${WD}"/flame.sh "${PERF_FILE}"

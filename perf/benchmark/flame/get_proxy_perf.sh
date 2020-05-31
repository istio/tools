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

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

function usage() {
  echo "usage:
        ./get_proxy_perf.sh -p <pod_name> -n <pod_namespace> -d <duration> -f <sample_frequency>
    
    e.g.
      ./get_proxy_perf.sh -p svc05-0-4-0-67bff5dbbf-grl94 -n service-graph05 -d 20 -f 99

    -p name of the pod.
    -n namespace of the given pod.
    -d time duration of profiling in second.
    -f sample frequency in Hz."
  exit 1
}

while getopts p:n:d:f: arg ; do
  case "${arg}" in
    p) POD_NAME="${OPTARG}";;
    n) POD_NAMESPACE="${OPTARG}";;
    d) PERF_DURATION="${OPTARG}";;
    f) SAMPLE_FREQUENCY="${OPTARG}";;
    *) usage;;
  esac
done

POD_NAME=${POD_NAME:?"pod name must be provided"}
POD_NAMESPACE=${POD_NAMESPACE:?"pod namespace must be provided"}
PERF_DURATION=${PERF_DURATION:-"20"}
SAMPLE_FREQUENCY=${SAMPLE_FREQUENCY:-"99"}
PERF_DATA_FILENAME=${PERF_DATA_FILENAME:-"perf.data"}

WD=$(dirname "${0}")
WD=$(cd "${WD}" && pwd)

echo "Copy profiling script to proxy..."
kubectl cp "${WD}"/get_perfdata.sh "${POD_NAME}":/etc/istio/proxy/get_perfdata.sh -n "${POD_NAMESPACE}" -c istio-proxy

echo "Start profiling..."
kubectl exec "${POD_NAME}" -n "${POD_NAMESPACE}" -c istio-proxy -- /etc/istio/proxy/get_perfdata.sh "${PERF_DATA_FILENAME}" "${PERF_DURATION}" "${SAMPLE_FREQUENCY}"

TMP_DIR=$(mktemp -d -t proxy-perf-XXXXXXXXXX)
trap 'rm -rf "${TMP_DIR}"' EXIT

if [[ "${PERF_DATA_FILENAME}" = "perf.data" ]]; then
    TIME="$(date '+%Y-%m-%d-%H-%M-%S')"
    PERF_FILE_NAME="${POD_NAME}_${TIME}.perf"
else
    PERF_FILE_NAME="${PERF_DATA_FILENAME}.perf"
fi

PERF_FILE="${TMP_DIR}/${PERF_FILE_NAME}"
kubectl cp "${POD_NAME}:/etc/istio/proxy/${PERF_DATA_FILENAME}.perf" "${PERF_FILE}" -n "${POD_NAMESPACE}" -c istio-proxy

echo "Generating svg file ${PERF_FILE_NAME}"
"${WD}/flame.sh" "${PERF_FILE}"

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
        ./get_istiod_pprof.sh -p <pod_name> -n <pod_namespace> -d <duration>

    e.g.
      ./get_istiod_pprof.sh -p istiod-67bff5dbbf-grl94 -n istio-system -d 20

    -p name of the pod.
    -n namespace of the given pod.
    -d time duration of profiling in second."
  exit 1
}

while getopts p:n:d:f: arg ; do
  case "${arg}" in
    p) POD_NAME="${OPTARG}";;
    n) POD_NAMESPACE="${OPTARG}";;
    d) PPROF_DURATION="${OPTARG}";;
    *) usage;;
  esac
done

POD_NAME=${POD_NAME:?"pod name must be provided"}
POD_NAMESPACE=${POD_NAMESPACE:?"pod namespace must be provided"}
PPROF_DURATION=${PPROF_DURATION:-"20"}
PPROF_DATA_FILENAME=${PPROF_DATA_FILENAME:-"profile.pprof"}

WD=$(dirname "${0}")
WD=$(cd "${WD}" && pwd)


TMP_DIR=$(mktemp -d -t istiod-pprof-XXXXXXXXXX)
trap 'rm -rf "${TMP_DIR}"' EXIT

if [[ "${PPROF_DATA_FILENAME}" = "profile.pprof" ]]; then
    TIME="$(date '+%Y-%m-%d-%H-%M-%S')"
    PPROF_FILENAME="${POD_NAME}_${TIME}.pprof"
else
    PPROF_FILENAME="${PPROF_DATA_FILENAME}"
fi

echo "Port forwarding debug port..."
kubectl -n "${POD_NAMESPACE}" port-forward "${POD_NAME}" 8080:8080 & PF_PID=${!}
# ensure port-forward is actually running
sleep 5; kill -0 "${PF_PID}"
trap 'kill ${PF_PID}' EXIT

echo "Start profiling..."
PERF_FILE="${TMP_DIR}/${PPROF_FILENAME}"
go tool pprof -seconds "${PPROF_DURATION}" -raw -output="${PERF_FILE}" http://127.0.0.1:8080/debug/pprof/profile
#
echo "Generating svg file ${PPROF_FILENAME}"
COLLAPSE_SCRIPT="go" "${WD}/flame.sh" "${PERF_FILE}"
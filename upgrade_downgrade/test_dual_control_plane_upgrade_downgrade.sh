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

set -x
set -o pipefail

WD=$(dirname "$0")
WD=$(cd "$WD" || exit; pwd)
ROOT=$(dirname "$WD")

command -v helm >/dev/null 2>&1 || { echo >&2 "helm must be installed, aborting."; exit 1; }

ISTIO_NAMESPACE="istio-system"
# Maximum % of 503 response that cannot exceed
MAX_503_PCT_FOR_PASS="15"
# Maximum % of connection refused that cannot exceed
# Set it to high value so it fails for explicit sidecar issues
MAX_CONNECTION_ERR_FOR_PASS="30"

while (( "$#" )); do
  PARAM=$(echo "${1}" | awk -F= '{print $1}')
  eval VALUE="$(echo "${1}" | awk -F= '{print $2}')"
  case "${PARAM}" in
    --namespace)
        ISTIO_NAMESPACE=${VALUE}
        ;;
    --from_hub)
        FROM_HUB=${VALUE}
        ;;
    --from_tag)
        FROM_TAG=${VALUE}
        ;;
    --from_path)
        FROM_PATH=${VALUE}
        ;;
    --to_hub)
        TO_HUB=${VALUE}
        ;;
    --to_tag)
        TO_TAG=${VALUE}
        ;;
    --to_path)
        TO_PATH=${VALUE}
        ;;
    --cloud)
        CLOUD=${VALUE}
        ;;
    *)
        echo "ERROR: unknown parameter \"$PARAM\""
        exit 1
        ;;
  esac
  shift
done

# Check if required parameters are passed
if [[ -z "${FROM_HUB}" || -z "${FROM_TAG}" || -z "${FROM_PATH}" || -z "${TO_HUB}" || -z "${TO_TAG}" || -z "${TO_PATH}" ]]; then
  echo "Error: from_hub, from_tag, from_path, to_hub, to_tag, to_path must all be set."
  exit 1
fi

# Check if scenario is a valid one
if [[ "${TEST_SCENARIO}" == "dual-control-plane-upgrade" || "${TEST_SCENARIO}" == "dual-control-plane-rollback" ]];then
  echo "The current test scenario is ${TEST_SCENARIO}."
else
  echo "Invalid scenario: ${TEST_SCENARIO}"
  echo "supported: dual-control-plane-upgrade, dual-control-plane-upgrade-downgrade"
fi

# shellcheck disable=SC1090
source "${ROOT}/upgrade_downgrade/common.sh"
# shellcheck disable=SC1090
source "${ROOT}/upgrade_downgrade/fortio_utils.sh"

# Check if istioctl is present in both "from" and "to" versions
FROM_ISTIOCTL="${FROM_PATH}/bin/istioctl"
TO_ISTIOCTL="${TO_PATH}/bin/istioctl"
if [[ ! -f "${FROM_ISTIOCTL}" || ! -f "${TO_ISTIOCTL}" ]]; then
  echo "istioctl not found in either ${FROM_PATH}/bin or ${TO_PATH}/bin directory"
  exit 1
fi

TMP_DIR=/tmp/istio_upgrade_test
POD_FORTIO_LOG=${TMP_DIR}/fortio_pod.log
TEST_NAMESPACE="test"

# Needed because --revision cannot have dots. That
# causes issues while installing. Also we need to truncate
# to 60 chars for this to work with 1.7 and below. But for
# our purposes we don't need those many.
TO_REVISION=$(echo "${TO_TAG}" | tr '.' '-' | cut -c -20)

user="cluster-admin"
if [[ "${CLOUD}" == "GKE" ]];then
  user="$(gcloud config get-value core/account)"
fi
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="${user}" || echo "clusterrolebinding already created."

write_msg "Reset cluster"
copy_test_files
reset_cluster "${TO_ISTIOCTL}"

write_msg "Deploy Istio(minimal) ${FROM_TAG}"
${FROM_ISTIOCTL} install -y --set profile=minimal
kubectl wait --for=condition=ready --timeout=10m pod --all -n "${ISTIO_NAMESPACE}"

write_msg "Deploy Echo v1 and v2"
kubectl apply -f "${TMP_DIR}/fortio.yaml" -n "${TEST_NAMESPACE}"
kubectl wait --for=condition=ready --timeout=10m pod --all -n "${TEST_NAMESPACE}"

write_msg "Generate internal traffic for echo v1 and v2"
kubectl apply -f "${TMP_DIR}/fortio-cli.yaml" -n "${TEST_NAMESPACE}"

# Install Istio 1.8 minimal profile with canary revision
write_msg "Deploy Istio(minimal) ${TO_TAG}"
${TO_ISTIOCTL} install -y --set profile=minimal --set revision="${TO_REVISION}"
kubectl wait --all --for=condition=Ready pods -n istio-system --timeout=5m

# Relabel namespace before restarting each service
write_msg "Relabel namespace to inject ${TO_TAG} proxy"
kubectl label namespace "${TEST_NAMESPACE}" istio-injection- istio.io/rev="${TO_REVISION}"

restart_data_plane echosrv-deployment-v1 "${TEST_NAMESPACE}"

if [[ "${TEST_SCENARIO}" == "dual-control-plane-upgrade" ]]; then
  restart_data_plane echosrv-deployment-v2 "${TEST_NAMESPACE}"
  write_msg "UPGRADE: Uninstall old version of control plane (${FROM_TAG})"
  PROFILE_YAML="${FROM_PATH}/manifests/profiles/minimal.yaml"
  uninstall_istio "${FROM_ISTIOCTL}" "${PROFILE_YAML}"

elif [[ "${TEST_SCENARIO}" == "dual-control-plane-rollback" ]]; then
  kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev- istio-injection=enabled
  restart_data_plane echosrv-deployment-v1 "${TEST_NAMESPACE}"
  write_msg "ROLLBACK: Uninstall new version of control plane (${TO_TAG})"
  uninstall_istio "${TO_ISTIOCTL}" "" "${TO_REVISION}"
fi

cli_pod_name=$(kubectl -n "${TEST_NAMESPACE}" get pods -lapp=cli-fortio -o jsonpath='{.items[0].metadata.name}')
kubectl wait --for=condition=complete --timeout=30m job/cli-fortio -n "${TEST_NAMESPACE}"

write_msg "Verify results"
kubectl logs -f -n "${TEST_NAMESPACE}" -c echosrv "${cli_pod_name}" &> "${POD_FORTIO_LOG}" || echo "Could not find ${cli_pod_name}"
if ! analyze_fortio_logs "${POD_FORTIO_LOG}" "${MAX_503_PCT_FOR_PASS}" "${MAX_CONNECTION_ERR_FOR_PASS}"; then
  failed=true
fi

if [[ -n "${failed}" ]]; then
  echo "FAILURE"
  exit 1
fi

echo "SUCCESS"

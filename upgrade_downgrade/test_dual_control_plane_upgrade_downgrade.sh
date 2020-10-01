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
SERVICE_UNAVAILABLE_CODE="503"
CONNECTION_ERROR_CODE="-1"

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


writeMsg "Reset cluster"
copy_test_files
resetCluster "${TO_ISTIOCTL}"

writeMsg "Deploy Istio(minimal) ${FROM_TAG}"
${FROM_ISTIOCTL} install -y --set profile=minimal
waitForPodsReady "${ISTIO_NAMESPACE}"

writeMsg "Deploy Echo v1 and v2"
kubectl apply -f "${TMP_DIR}/fortio.yaml" -n "${TEST_NAMESPACE}"
waitForPodsReady "${TEST_NAMESPACE}"

writeMsg "Generate internal traffic for echo v1 and v2"
kubectl apply -f "${TMP_DIR}/fortio-cli.yaml" -n "${TEST_NAMESPACE}"

# Install Istio 1.8 minimal profile with canary revision
writeMsg "Deploy Istio(minimal) ${TO_TAG}"
${TO_ISTIOCTL} install -y --set profile=minimal --set revision="${TO_REVISION}"
kubectl wait --all --for=condition=Ready pods -n istio-system --timeout=5m

# Relabel namespace before restarting each service
writeMsg "Relabel namespace to inject ${TO_TAG} proxy"
kubectl label namespace "${TEST_NAMESPACE}" istio-injection- istio.io/rev="${TO_REVISION}"

function verifyIstiod() {
  local ns="$1"
  local app="$2"
  local version="$3"
  local istioctl_path="$4"
  local expected="$5"

  local mismatch=0

  for pod in $(kubectl get pod -lapp="$app" -lversion="$version" -n "$ns" -o name); do
    local istiod
    local podname
    podname=$(echo "$pod" | cut -d'/' -f2)
    istiod=$(${istioctl_path} proxy-config endpoint "$podname.$ns" --cluster xds-grpc -o json | jq -r '.[].hostStatuses[].hostname')
    echo "  $pod ==> ${istiod}"
    if [[ "$istiod" != *"$expected"* ]]; then
      mismatch=$(( mismatch+1 ))
    fi
  done

  if ((mismatch == 0)); then
    return 0
  fi
  return 1
}

kubectl rollout restart deployment echosrv-deployment-v1 -n "${TEST_NAMESPACE}"
withRetries 30 10 checkDeploymentRolledOut "${TEST_NAMESPACE}" echosrv-deployment-v1
withRetries 5 20 verifyIstiod "${TEST_NAMESPACE}" "echosrv-deployment-v1" "v1" \
  "${TO_ISTIOCTL}" "istiod-${TO_REVISION}.istio-system.svc"
withRetries 5 20 verifyIstiod "${TEST_NAMESPACE}" "echosrv-deployment-v2" "v2" \
  "${TO_ISTIOCTL}" "istiod.istio-system.svc"

if [[ "${TEST_SCENARIO}" == "dual-control-plane-upgrade" ]]; then
  kubectl rollout restart deployment echosrv-deployment-v2 -n "${TEST_NAMESPACE}"
  withRetries 30 10 checkDeploymentRolledOut "${TEST_NAMESPACE}" "echosrv-deployment-v2"
  withRetries 5 20 verifyIstiod "${TEST_NAMESPACE}" "echosrv-deployment-v2" "v2" \
    "${TO_ISTIOCTL}" "istiod-${TO_REVISION}.istio-system.svc"
  writeMsg "UPGRADE: Uninstall old version of control plane (${FROM_TAG})"
  ${FROM_ISTIOCTL} experimental uninstall --filename -y "${FROM_PATH}/manifests/profiles/minimal.yaml"

elif [[ "${TEST_SCENARIO}" == "dual-control-plane-rollback" ]]; then
  kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev- istio-injection=enabled
  kubectl rollout restart deployment echosrv-deployment-v1 -n "${TEST_NAMESPACE}"
  withRetries 30 10 checkDeploymentRolledOut "${TEST_NAMESPACE}" "echosrv-deployment-v1"
  withRetries 5 20 verifyIstiod "${TEST_NAMESPACE}" "echosrv-deployment-v1" "v1" \
    "${TO_ISTIOCTL}" "istiod.istio-system.svc"
  writeMsg "ROLLBACK: Uninstall new version of control plane (${TO_TAG})"
  ${TO_ISTIOCTL} experimental uninstall --revision "${TO_REVISION}" -y
fi

cli_pod_name=$(kubectl -n "${TEST_NAMESPACE}" get pods -lapp=cli-fortio -o jsonpath='{.items[0].metadata.name}')
waitForJob cli-fortio "${TEST_NAMESPACE}"

writeMsg "Verify results"
kubectl logs -f -n "${TEST_NAMESPACE}" -c echosrv "${cli_pod_name}" &> "${POD_FORTIO_LOG}" || echo "Could not find ${cli_pod_name}"
pod_log_str=$(grep "Code 200"  "${POD_FORTIO_LOG}")

cat ${POD_FORTIO_LOG}

if [[ ${pod_log_str} != *"Code 200"* ]];then
  echo "=== No Code 200 found in internal traffic log ==="
  failed=true
elif ! errorPercentBelow "${POD_FORTIO_LOG}" "${SERVICE_UNAVAILABLE_CODE}" ${MAX_503_PCT_FOR_PASS}; then
  echo "=== Code 503 Errors found in internal traffic exceeded ${MAX_503_PCT_FOR_PASS}% threshold ==="
  failed=true
elif ! errorPercentBelow "${POD_FORTIO_LOG}" "${CONNECTION_ERROR_CODE}" ${MAX_CONNECTION_ERR_FOR_PASS}; then
  echo "=== Connection Errors found in internal traffic exceeded ${MAX_CONNECTION_ERR_FOR_PASS}% threshold ==="
  failed=true
else
  echo "=== Errors found in internal traffic is within threshold ==="
fi

if [[ -n "${failed}" ]]; then
  echo "FAILURE"
  exit 1
fi

echo "SUCCESS"

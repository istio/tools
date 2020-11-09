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

command -v fortio >/dev/null 2>&1 || { echo >&2 "fortio must be installed, aborting."; exit 1; }

ISTIO_NAMESPACE="istio-system"
MAX_5XX_PCT_FOR_PASS="15"
# Maximum % of connection refused that cannot exceed
# Set it to high value so it fails for explicit sidecar issues
MAX_503_PCT_FOR_PASS="15"
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
if [[ "${TEST_SCENARIO}" == "boutique-upgrade" || "${TEST_SCENARIO}" == "boutique-rollback" ]];then
  echo "The current test scenario is ${TEST_SCENARIO}."
else
  echo "Invalid scenario: ${TEST_SCENARIO}"
  echo "supported: boutique-upgrade, boutique-rollback"
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
FROM_REVISION=$(echo "${FROM_TAG}" | tr '.' '-' | cut -c -20)
TO_REVISION=$(echo "${TO_TAG}" | tr '.' '-' | cut -c -20)

user="cluster-admin"
if [[ "${CLOUD}" == "GKE" ]];then
  user="$(gcloud config get-value core/account)"
fi
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="${user}" || echo "clusterrolebinding already created."

writeMsg "Reset cluster"
copy_test_files
resetCluster "${TO_ISTIOCTL}"

# Clone the repository
REPO_PATH=$(mktemp -d "microservices-demo-XXXXX")
git clone https://github.com/GoogleCloudPlatform/microservices-demo.git "${REPO_PATH}" --depth=1

# Install Initial version of Istio
writeMsg "Deploy Istio ${FROM_TAG}"
${FROM_ISTIOCTL} install -y --revision "${FROM_REVISION}" 
waitForPodsReady "${ISTIO_NAMESPACE}"

# 1. Create namespace and label for automatic injection
# 2. Deploy online boutique application and Istio configuration
kubectl create namespace "${TEST_NAMESPACE}"
kubectl label namespace "${TEST_NAMESPACE}" istio-injection- || echo "istio-injection label removed"
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev="${FROM_REVISION}"

# It is not as simple as throwing YAML at it. Problem is, there are dependencies between
# services and loadgenerator should not start before others as it will fall into CrashLoopBackoff
# So we have to make sure that if something has fallen into CrashLoopBackoff then keep restarting
# the deployment so that the timeout wouldn't be too long.
function deploy_boutique_shop_app() {
  local test_ns="$1"
  kubectl apply -f "${REPO_PATH}/release/kubernetes-manifests.yaml" -n "${test_ns}"
  while true; do
    local pod_count=0
    local run_count=0
    for p in $(kubectl get pods -n "${test_ns}" -o name); do
      pod_count=$((pod_count+1))
      local pod_line=$(kubectl get "$p" -n "${test_ns}")
      if [[ "${pod_line}" == *"Running"* ]]; then
        run_count=$((run_count+1))
        continue
      fi
      if [[ "$pod_line" == *CrashLoopBackOff* || "$pod_line" == *Error* ]]; then
        # NOTE: Boutique shop app deployment YAML has a specific pattern
        # where name of the deployment matches the value of app label. That
        # is app=<xyz> would have its deployment name as <xyz>
        local deployment_name=$(kubectl get "$p" -n "${test_ns}" -o jsonpath='{.metadata.labels}' | jq -r 'app')
        echo "$p is stuck in CrashLoopBackoff or Error. So restart deployment ${deployment_name}"
        kubectl rollout restart deployment "${deployment_name}" -n "${test_ns}"
      fi
    done
    if (( run_count == pod_count )); then
      echo "Boutique shop deployed successfully"
      break
    fi
    sleep 10
  done
}

deploy_boutique_shop_app "${TEST_NAMESPACE}"
kubectl apply -f "${REPO_PATH}/release/istio-manifests.yaml" -n "${TEST_NAMESPACE}"

# But load-generator should not have a sidecar. So patch its deployment
# and restart deployment
kubectl patch deployment loadgenerator -n "${TEST_NAMESPACE}" --patch '{"spec":{"template":{"metadata":{"annotations": {"sidecar.istio.io/inject": "false"}}}}}'
kubectl rollout restart deployment loadgenerator -n "${TEST_NAMESPACE}"
waitForPodsReady "${TEST_NAMESPACE}"

# Start external traffic from fortio
# 1. First get ingress address
# 2. Next, use that address to fire requests at boutique shop app
waitForIngress

TRAFFIC_RUNTIME_SEC=800
LOCAL_FORTIO_LOG=${TMP_DIR}/fortio_local.log
EXTERNAL_FORTIO_DONE_FILE=${TMP_DIR}/fortio_done_file

# TODO(su225): Move it to common as it is generic enoug
# or probably fortio_helper.sh?
runFortioLoadCommand() {
  withRetries 10 10  echo_and_run fortio load -c 32 -t "${TRAFFIC_RUNTIME_SEC}"s -qps 10 -timeout 30s\
    "http://${1}/" &> "${LOCAL_FORTIO_LOG}"
  echo "done" >> "${EXTERNAL_FORTIO_DONE_FILE}"
}

waitForExternalRequestTraffic() {
  echo "Waiting for external traffic to complete"
  local attempt=0
  while [[ ! -f "${EXTERNAL_FORTIO_DONE_FILE}" ]]; do
    echo "attempt ${attempt}"
    attempt=$((attempt+1))
    sleep 10
  done
}

# Sends external traffic from machine test is running on through external IP and ingress gateway LB.
sendExternalRequestTraffic() {
  writeMsg "Sending external traffic"
  runFortioLoadCommand "${1}"
}

# Start sending traffic from outside the cluster
sendExternalRequestTraffic "${INGRESS_ADDR}" &

# Wait for some time
# Represents stabilizing period
STABILIZING_PERIOD=60
writeMsg "Wait for $STABILIZING_PERIOD seconds"
sleep "${STABILIZING_PERIOD}"

# Install Target revision of Istio
writeMsg "Deploy Istio ${TO_TAG}"
${TO_ISTIOCTL} install -y --revision "${TO_REVISION}"
waitForPodsReady "${ISTIO_NAMESPACE}"

# Now change labels and restart deployments one at a time
# But **DO NOT** restart loadgenerator. We need that to run
# continuously during upgrade so that we can see how many
# requests fail during data plane restart after upgrade.
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev-
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev="${TO_REVISION}"
for d in $(kubectl get deployments -n "${TEST_NAMESPACE}" -o name | grep -v loadgenerator); do
  deployment=$(echo "$d" | cut -d'/' -f2)
  restartDataPlane "$deployment" "${TEST_NAMESPACE}"
done

waitForPodsReady "${TEST_NAMESPACE}"

for d in $(kubectl get deployments -n "${TEST_NAMESPACE}" -o name | grep -v loadgenerator); do
  app_label=$(kubectl get deployment "$deployment" -n "${TEST_NAMESPACE}" -o jsonpath='{.spec.selector.matchLabels.app}')  
  for pod in $(kubectl get pods -lapp="${app_label}" -n "${TEST_NAMESPACE}" -o name); do
    istiod=$(getIstiod "${TO_ISTIOCTL}" "${pod}" "${TEST_NAMESPACE}")
    expected_istiod="istiod-${TO_REVISION}.${ISTIO_NAMESPACE}"
    if [[ "$istiod" != *"${expected_istiod}"* ]]; then
      echo "$pod is not pointing to right istiod. Expected **$expected_istiod**, but got $istiod"
      exit 1
    fi
  done
done

waitForExternalRequestTraffic

# Finally look at the statistics and check failure percentages
aggregated_stats="$(kubectl logs $(kubectl get pods -lapp=loadgenerator -n ${TEST_NAMESPACE} -o name) -n ${TEST_NAMESPACE} | grep 'Aggregated' | tail -1)"
echo "$aggregated_stats"

internal_failure_percent="$(echo $aggregated_stats | awk '{ print $3 }' | tr '()%' ' ' | cut -d' ' -f2)"
if [[ $(python -c "print($internal_failure_percent > ${MAX_503_PCT_FOR_PASS})") == *True* ]]; then
  failed=true
fi

# Now get fortio logs for the process running outside cluster
local_log_str=$(grep "Code 200" "${LOCAL_FORTIO_LOG}")
cat ${LOCAL_FORTIO_LOG}
if [[ ${local_log_str} != *"Code 200"* ]];then
  echo "=== No Code 200 found in external traffic log ==="
  failed=true
elif ! errorPercentBelow "${LOCAL_FORTIO_LOG}" "${SERVICE_UNAVAILABLE_CODE}" ${MAX_503_PCT_FOR_PASS}; then
  echo "=== Code 503 Errors found in external traffic exceeded ${MAX_503_PCT_FOR_PASS}% threshold ==="
  failed=true
elif ! errorPercentBelow "${LOCAL_FORTIO_LOG}" "${CONNECTION_ERROR_CODE}" ${MAX_CONNECTION_ERR_FOR_PASS}; then
  echo "=== Connection Errors found in external traffic exceeded ${MAX_CONNECTION_ERR_FOR_PASS}% threshold ==="
  failed=true
else
  echo "=== Errors found in external traffic is within threshold ==="
fi

if [[ -n "${failed}" ]]; then
  exit 1
fi

echo "SUCCESS"

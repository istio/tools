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
TEST_NAMESPACE="test"
FROM_REVISION=$(echo "${FROM_TAG}" | tr '.' '-' | cut -c -20)
TO_REVISION=$(echo "${TO_TAG}" | tr '.' '-' | cut -c -20)

user="cluster-admin"
if [[ "${CLOUD}" == "GKE" ]];then
  user="$(gcloud config get-value core/account)"
fi
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="${user}" || echo "clusterrolebinding already created."

write_msg "Reset cluster"
copy_test_files
reset_cluster "${TO_ISTIOCTL}"

# Install Initial version of Istio
write_msg "Deploy Istio ${FROM_TAG}"
${FROM_ISTIOCTL} install -f "${TMP_DIR}/iop-control-plane.yaml" -y --revision "${FROM_REVISION}" || die "control plane installation failed"
${FROM_ISTIOCTL} install -f "${TMP_DIR}/iop-gateways.yaml" -y --revision "${FROM_REVISION}" || die "gateway installation failed"
wait_for_pods_ready "${ISTIO_NAMESPACE}"

# 1. Create namespace and label for automatic injection
# 2. Deploy online boutique application and Istio configuration
kubectl get namespace "${TEST_NAMESPACE}" || kubectl create namespace "${TEST_NAMESPACE}"
kubectl label namespace "${TEST_NAMESPACE}" istio-injection- || echo "istio-injection label removed"
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev="${FROM_REVISION}"

# It is not as simple as throwing YAML at it. Problem is, there are dependencies between
# services and loadgenerator should not start before others as it will fall into CrashLoopBackoff
# So we have to make sure that if something has fallen into CrashLoopBackoff then keep restarting
# the deployment so that the timeout wouldn't be too long.
function deploy_boutique_shop_app() {
  local test_ns="$1"
  kubectl apply -f "https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/master/release/kubernetes-manifests.yaml" -n "${test_ns}"
  while true; do
    local pod_count=0
    local run_count=0
    for p in $(kubectl get pods -n "${test_ns}" -o name); do
      pod_count=$((pod_count+1))
      local pod_line
      pod_line=$(kubectl get "$p" -n "${test_ns}")
      if [[ "${pod_line}" == *"Running"* ]]; then
        run_count=$((run_count+1))
        continue
      fi
      if [[ "$pod_line" == *CrashLoopBackOff* || "$pod_line" == *Error* ]]; then
        # NOTE: Boutique shop app deployment YAML has a specific pattern
        # where name of the deployment matches the value of app label. That
        # is app=<xyz> would have its deployment name as <xyz>
        local deployment_name
        deployment_name=$(kubectl get "$p" -n "${test_ns}" -o jsonpath='{.metadata.labels}' | jq -r '.app')
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
kubectl apply -f "https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/master/release/istio-manifests.yaml" -n "${TEST_NAMESPACE}"

# But load-generator should not have a sidecar. So patch its deployment
# and restart deployment
kubectl patch deployment loadgenerator -n "${TEST_NAMESPACE}" --patch '{"spec":{"template":{"metadata":{"annotations": {"sidecar.istio.io/inject": "false"}}}}}'
kubectl rollout restart deployment loadgenerator -n "${TEST_NAMESPACE}"
wait_for_pods_ready "${TEST_NAMESPACE}"

# Start external traffic from fortio
# 1. First get ingress address
# 2. Next, use that address to fire requests at boutique shop app
wait_for_ingress

export TRAFFIC_RUNTIME_SEC=800
export LOCAL_FORTIO_LOG=${TMP_DIR}/fortio_local.log
export EXTERNAL_FORTIO_DONE_FILE=${TMP_DIR}/fortio_done_file

# Start sending traffic from outside the cluster
send_external_request_traffic "http://${INGRESS_ADDR}" &

# Wait for some time
# Represents stabilizing period
STABILIZING_PERIOD=60
write_msg "Wait for $STABILIZING_PERIOD seconds"
sleep "${STABILIZING_PERIOD}"

# Install Target revision of Istio
write_msg "Deploy Istio ${TO_TAG}"
${TO_ISTIOCTL} install -f "${TMP_DIR}/iop-control-plane.yaml" -y --revision "${TO_REVISION}" || die "installing ${TO_REVISION} control plane failed"
${TO_ISTIOCTL} install -f "${TMP_DIR}/iop-gateways.yaml" -y --revision "${TO_REVISION}" || die "installing ${TO_REVISION} gateways failed"
wait_for_pods_ready "${ISTIO_NAMESPACE}"

# Now change labels and restart deployments one at a time
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev-
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev="${TO_REVISION}"

# Upgrade select micro-services in the first round
first_round=( "frontend" "redis-cart" "paymentservice" )
for d in "${first_round[@]}"; do
  restart_data_plane "$d" "${TEST_NAMESPACE}"
done

# **DO NOT** restart loadgenerator. We need that to run
# continuously during upgrade so that we can see how many
# requests fail during data plane restart after upgrade.
if [[ "${TEST_SCENARIO}" == "boutique-upgrade" ]]; then
  for d in $(kubectl get deployments -n "${TEST_NAMESPACE}" -o name | grep -v loadgenerator); do
    deployment=$(echo "$d" | cut -d'/' -f2)
    if [[ ! "${first_round[*]}" =~ $deployment ]]; then
      restart_data_plane "$deployment" "${TEST_NAMESPACE}"
    fi
  done
fi

wait_for_pods_ready "${TEST_NAMESPACE}"

if [[ "${TEST_SCENARIO}" == "boutique-upgrade" ]]; then
  write_msg "uninstalling ${FROM_REVISION}"
  # Currently we don't do it for gateways because gateway upgrade is still in-place :(
  # So remove only control plane with old revision.
  "${FROM_ISTIOCTL}" experimental uninstall -f "${TMP_DIR}/iop-control-plane.yaml" --revision "${FROM_REVISION}" -y || die "uninstalling control plane ${FROM REVISION} failed"
else
  kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev-
  kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev="${FROM_REVISION}"

  for d in "${first_round[@]}"; do
    restart_data_plane "$d" "${TEST_NAMESPACE}"
  done
  wait_for_pods_ready "${TEST_NAMESPACE}"

  write_msg "uninstalling ${TO_REVISION}"
  "${TO_ISTIOCTL}" experimental uninstall -f "${TMP_DIR}/iop-control-plane.yaml" --revision "${TO_REVISION}" -y || die "uninstalling control plane ${TO_REVISION} failed"
  "${FROM_ISTIOCTL}" install -f "${TMP_DIR}/iop-gateways.yaml" --revision "${FROM_REVISION}" -y || die "installing ${FROM_REVISION} gateways failed"
fi

wait_for_external_request_traffic

# Finally look at the statistics and check failure percentages
MAX_5XX_PCT_FOR_PASS="15"
MAX_CONNECTION_ERR_FOR_PASS="30"
loadgen_pod="$(kubectl get pods -lapp=loadgenerator -n ${TEST_NAMESPACE} -o name)"
aggregated_stats=$(kubectl logs "${loadgen_pod}"  -n ${TEST_NAMESPACE} | grep 'Aggregated' | tail -1)
echo "$aggregated_stats"

internal_failure_percent=$(echo "${aggregated_stats}" | awk '{ print $3 }' | tr '()%' ' ' | cut -d' ' -f2)
if ! cmp_float_le "${internal_failure_percent}" "${MAX_5XX_PCT_FOR_PASS}"; then
  failed=true
fi

# Now get fortio logs for the process running outside cluster
write_msg "Analyze external fortio log file ${LOCAL_FORTIO_LOG}"
if ! analyze_fortio_logs "${LOCAL_FORTIO_LOG}" "${MAX_5XX_PCT_FOR_PASS}" "${MAX_CONNECTION_ERR_FOR_PASS}"; then
  failed=true
fi

if [[ -n "${failed}" ]]; then
  exit 1
fi

echo "SUCCESS"

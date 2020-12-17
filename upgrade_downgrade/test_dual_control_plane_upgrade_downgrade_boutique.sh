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
LOADGEN_NAMESPACE="loadgen"
FROM_REVISION=$(echo "${FROM_TAG}" | tr '.' '-' | cut -c -20)
TO_REVISION=$(echo "${TO_TAG}" | tr '.' '-' | cut -c -20)

export TRAFFIC_RUNTIME_SEC=800
export LOCAL_FORTIO_LOG=${TMP_DIR}/fortio_local.log
export EXTERNAL_FORTIO_DONE_FILE=${TMP_DIR}/fortio_done_file

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
kubectl wait --for=condition=ready pod --all -n "${ISTIO_NAMESPACE}" --timeout=30m

# Create namespace and label for automatic injection
kubectl get namespace "${TEST_NAMESPACE}" || kubectl create namespace "${TEST_NAMESPACE}"
kubectl get namespace "${LOADGEN_NAMESPACE}" || kubectl create namespace "${LOADGEN_NAMESPACE}"
kubectl label namespace "${TEST_NAMESPACE}" istio-injection- || echo "istio-injection label removed"
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev="${FROM_REVISION}"

function deploy_and_wait_for_services() {
  # **WARNING**: Do not name this services. We'll end up with a situation
  # which looks like "local -n services=services". Bash complains
  # that this is a circular reference. This resulted in test failures.
  #
  # In general, we should not end up in a situation like this
  # local -n svcs=svcs (happens when array passed is also named svcs)
  local -n svcs="$1"
  for svc in "${svcs[@]}"; do
    local service_manifest_file="${TMP_DIR}/boutique/k8s-${svc}.yaml"
    if [[ ! -f "${service_manifest_file}" ]]; then
      echo "Failed to deploy: Kubernetes manifest file for ${svc} not found - ${service_manifest_file}"
      return 1
    fi
    kubectl apply -f "${service_manifest_file}" # Namespace is already defined in manifest
  done
  kubectl wait --for=condition=ready pod --all -n "${TEST_NAMESPACE}" --timeout=30m
}

# Setup boutique shop app. Namespace is specified in YAML.
# Link to architecture diagram:
# https://github.com/GoogleCloudPlatform/microservices-demo/blob/master/docs/img/architecture-diagram.png 

# Shellcheck is falsely complaining that these are not being used.
# shellcheck disable=SC2034
base_services=( "redis-cart" \
                "productcatalogservice" \
                "currencyservice" \
                "shippingservice" \
                "paymentservice" \
                "emailservice" \
                "adservice" )

# shellcheck disable=SC2034
services=( "cartservice" \
           "checkoutservice" \
           "recommendationservice" \
           "frontend" )

if ! deploy_and_wait_for_services base_services; then exit 1; fi
if ! deploy_and_wait_for_services services; then exit 1; fi

kubectl apply -f "${TMP_DIR}/boutique/istio-manifests.yaml"
kubectl apply -f "${TMP_DIR}/boutique/k8s-loadgenerator.yaml"
kubectl wait --for=condition=ready pod -n "${LOADGEN_NAMESPACE}" --timeout=30m

# Start external traffic from fortio
# 1. First get ingress address
# 2. Next, use that address to fire requests at boutique shop app
wait_for_ingress

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
kubectl wait --for=condition=ready pod --all -n "${ISTIO_NAMESPACE}" --timeout=30m

# Now change labels and restart deployments one at a time
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev-
kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev="${TO_REVISION}"

# Upgrade select micro-services in the first round
first_round=( "frontend" "redis-cart" "paymentservice" )
function restart_first_round() {
  for d in "${first_round[@]}"; do
    restart_data_plane "$d" "${TEST_NAMESPACE}"
  done
}

restart_first_round

if [[ "${TEST_SCENARIO}" == "boutique-upgrade" ]]; then
  for d in $(kubectl get deployments -n "${TEST_NAMESPACE}" -o name); do
    deployment=$(echo "$d" | cut -d'/' -f2)
    # If it is restarted in the first round, then don't restart again.
    if [[ ! "${first_round[*]}" =~ $deployment ]]; then
      restart_data_plane "$deployment" "${TEST_NAMESPACE}"
    fi
  done
fi

kubectl wait --for=condition=ready pod --all -n "${TEST_NAMESPACE}" --timeout=30m

if [[ "${TEST_SCENARIO}" == "boutique-upgrade" ]]; then
  write_msg "uninstalling ${FROM_REVISION}"
  # Currently we don't do it for gateways because gateway upgrade is still in-place :(
  # So remove only control plane with old revision.
  uninstall_istio "${FROM_ISTIOCTL}" "${TMP_DIR}/iop-control-plane.yaml" "${FROM_REVISION}" || \
    die "uninstalling control plane ${FROM REVISION} failed"
else
  kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev-
  kubectl label namespace "${TEST_NAMESPACE}" istio.io/rev="${FROM_REVISION}"

  restart_first_round
  kubectl wait --for=condition=ready pod --all -n "${TEST_NAMESPACE}" --timeout=30m

  write_msg "uninstalling ${TO_REVISION}"
  "${TO_ISTIOCTL}" experimental uninstall -f "${TMP_DIR}/iop-control-plane.yaml" --revision "${TO_REVISION}" -y
  uninstall_istio "${TO_ISTIOCTL}" "${TMP_DIR}/iop-control-plane.yaml" "${TO_REVISION}" || \
    die "uninstalling control plane ${TO_REVISION} failed"
  "${FROM_ISTIOCTL}" install -f "${TMP_DIR}/iop-gateways.yaml" --revision "${FROM_REVISION}" -y || die "installing ${FROM_REVISION} gateways failed"
fi

wait_for_external_request_traffic

# Finally look at the statistics and check failure percentages
loadgen_pod="$(kubectl get pods -lapp=loadgenerator -n ${LOADGEN_NAMESPACE} -o name)"
aggregated_stats=$(kubectl logs "${loadgen_pod}"  -n ${LOADGEN_NAMESPACE} | grep 'Aggregated' | tail -1)
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

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

# This test checks for control and data plane crossgrade. It runs the following steps:
# 1. Installs istio with multiple gateway replicas at from_hub/tag/path (path must point to a dir with release).
# 2. Installs fortio echosrv with a couple of different subsets/destination rules with multiple replicas.
# 3. Sends external traffic to echosrv through ingress gateway.
# 4. Sends internal traffic to echosrv from fortio load pod.
# 5. Upgrades control plane to to_hub/tag/path.
# 6. Does rolling restart of one of the echosrv subsets, which auto-injects upgraded version of sidecar.
# 7. Waits a while, then does downgrade.
# 8. Downgrades data plane by applying a saved ConfigMap from the previous version and doing rolling restart.
# 9. Downgrades control plane to from_hub/tag/path.
# 10. Parses the output from the load pod to check for any errors during upgrade and returns 1 is any are found.
#
# Dependencies that must be preinstalled: helm, fortio.
#

set -o pipefail

WD=$(dirname "$0")
WD=$(cd "$WD" || exit; pwd)
ROOT=$(dirname "$WD")

command -v helm >/dev/null 2>&1 || { echo >&2 "helm must be installed, aborting."; exit 1; }
command -v fortio >/dev/null 2>&1 || { echo >&2 "fortio must be installed, aborting."; exit 1; }

usage() {
    echo "Usage:"
    echo "  ./test_upgrade_downgrade.sh [OPTIONS]"
    echo
    echo "  from_hub          hub of release to upgrade from (required)."
    echo "  from_tag          tag of release to upgrade from (required)."
    echo "  from_path         path to release dir to upgrade from (required)."
    echo "  to_hub            hub of release to upgrade to (required)."
    echo "  to_tag            tag of release to upgrade to (required)."
    echo "  to_path           path to release to upgrade to (required)."
    echo "  install_options   install istio using either helm or istioctl (required)."
    echo "  auth_enable       enable mtls."
    echo "  skip_cleanup      leave install intact after test completes."
    echo "  namespace         namespace to install istio control plane in (default istio-system)."
    echo "  cloud             cloud provider name (required)"
    echo
    echo "  e.g. ./test_upgrade_downgrade.sh \"
    echo "        --from_hub=gcr.io/istio-testing --from_tag=d639408fd --from_path=/tmp/release-d639408fd \"
    echo "        --to_hub=gcr.io/istio-release --to_tag=1.0.2 --to_path=/tmp/istio-1.0.2 --cloud=GKE"
    echo "        --install_options=istioctl"
    echo
    exit 1
}

ISTIO_NAMESPACE="istio-system"
# Maximum % of 503 response that cannot exceed
MAX_503_PCT_FOR_PASS="15"
# Maximum % of connection refused that cannot exceed
# Set it to high value so it fails for explicit sidecar issues
MAX_CONNECTION_ERR_FOR_PASS="30"
SERVICE_UNAVAILABLE_CODE="503"
CONNECTION_ERROR_CODE="-1"

# TODO: later on, we add one more flag about supporting user specify the profile yaml file for upgrade
# Currently, we are supporting the default profiles
while (( "$#" )); do
    PARAM=$(echo "${1}" | awk -F= '{print $1}')
    eval VALUE="$(echo "${1}" | awk -F= '{print $2}')"
    case "${PARAM}" in
        -h | --help)
            usage
            exit
            ;;
        --namespace)
            ISTIO_NAMESPACE=${VALUE}
            ;;
        --skip_cleanup)
            SKIP_CLEANUP=true
            ;;
        --auth_enable)
            AUTH_ENABLE=true
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
        --install_options)
            INSTALL_OPTIONS=${VALUE}
            ;;
        --cloud)
            CLOUD=${VALUE}
            ;;
        *)
            echo "ERROR: unknown parameter \"$PARAM\""
            usage
            exit 1
            ;;
    esac
    shift
done

if [[ -z "${FROM_HUB}" || -z "${FROM_TAG}" || -z "${FROM_PATH}" || -z "${TO_HUB}" || -z "${TO_TAG}" || -z "${TO_PATH}" || -z "${INSTALL_OPTIONS}" ]]; then
    echo "Error: from_hub, from_tag, from_path, to_hub, to_tag, to_path, install_options must all be set."
    exit 1
fi

if [[ "${TEST_SCENARIO}" == "upgrade-downgrade" || "${TEST_SCENARIO}" == "upgrade" || "${TEST_SCENARIO}" == "downgrade" ]];then
    echo "The current test scenario is ${TEST_SCENARIO}."
else
    echo "Error: invalid test scenario: ${TEST_SCENARIO}."
    exit 1
fi

echo "Testing upgrade/downgrade from ${FROM_HUB}:${FROM_TAG} at ${FROM_PATH} to ${TO_HUB}:${TO_TAG} at ${TO_PATH} in namespace ${ISTIO_NAMESPACE}, auth=${AUTH_ENABLE}, cleanup=${SKIP_CLEANUP}"

TMP_DIR=/tmp/istio_upgrade_test
LOCAL_FORTIO_LOG=${TMP_DIR}/fortio_local.log
POD_FORTIO_LOG=${TMP_DIR}/fortio_pod.log

# Make sure to change templates/*.yaml with the correct address if this changes.
TEST_NAMESPACE="test"

# This must be at least as long as the script execution time.
# Edit fortio-cli.yaml to the same value when changing this.
TRAFFIC_RUNTIME_SEC=500

# Used to signal that background external process is done.
EXTERNAL_FORTIO_DONE_FILE=${TMP_DIR}/fortio_done_file

# shellcheck disable=SC1090
source "${ROOT}/upgrade_downgrade/common.sh"

installIstioAtVersionUsingHelm() {
    writeMsg "helm templating then applying new yaml using version ${2} from ${3}."
    if [[ -n "${AUTH_ENABLE}" ]]; then
        echo "Auth is enabled, generating manifest with auth."
        auth_opts="--set global.mtls.enabled=true"
    fi
    release_path="${3}"/install/kubernetes/helm/istio

    # See https://preliminary.istio.io/docs/setup/kubernetes/helm-install/
    helm init --client-only
    for i in "${3}"/install/kubernetes/helm/istio-init/files/crd*yaml; do
        echo_and_run kubectl apply -f "${i}"
    done
    sleep 5 # Per official Istio documentation!

    helm template "${release_path}" "${auth_opts}" \
    --name istio --namespace "${ISTIO_NAMESPACE}" \
    --set gateways.istio-ingressgateway.autoscaleMin=4 \
    --set pilot.autoscaleMin=2 \
    --set mixer.telemetry.autoscaleMin=2 \
    --set mixer.policy.autoscaleMin=2 \
    --set global.proxy.accessLogFile="/dev/stdout" \
    --set prometheus.enabled=false \
    --set-string global.hub="${1}" \
    --set-string global.tag="${2}" \
    --set global.defaultPodDisruptionBudget.enabled=true > "${WD}/istio.yaml" || die "helm template failed"

    kubectl apply -f "${WD}"/istio.yaml
}

installIstioAtVersionUsingIstioctl(){
  writeMsg "istioctl install istio using version ${2} from ${3}."
  istioctl_path="${3}"/bin
  find "${istioctl_path}" -maxdepth 1 -type f
  # shellcheck disable=SC2072
  if [[ "${2}" > "1.6" ]]; then
      "${istioctl_path}"/istioctl install --skip-confirmation
  elif [[ "${2}" > "1.4" ]]; then
      "${istioctl_path}"/istioctl manifest apply --skip-confirmation
  else
      "${istioctl_path}"/istioctl x manifest apply --yes
  fi
}

upgradeIstioAtVersionUsingIstioctl(){
  writeMsg "istioctl upgrade istio using version ${2} from ${3}."
  istioctl_path="${3}"/bin
  # shellcheck disable=SC2072
  if [[ "${TO_TAG}" > "1.6" ]]; then
    "${istioctl_path}"/istioctl upgrade --skip-confirmation --charts "${3}"/manifests
  elif [[ "${TO_TAG}" > "1.5" ]]; then
    # The following is a workaround for the test failure due to that the charts is
    # no longer bundled in the release (https://github.com/istio/istio/issues/23172).
    "${istioctl_path}"/istioctl manifest apply --skip-confirmation --charts "${3}"/manifests
  elif [[ "${TO_TAG}" > "1.4" ]]; then
    "${istioctl_path}"/istioctl x upgrade --skip-confirmation
  fi
}

istioInstallOptions() {
  if [[ "${INSTALL_OPTIONS}" == "helm" ]];then
    installIstioAtVersionUsingHelm "${FROM_HUB}" "${FROM_TAG}" "${FROM_PATH}"
  elif [[ "${INSTALL_OPTIONS}" == "istioctl" ]];then
    installIstioAtVersionUsingIstioctl "${FROM_HUB}" "${FROM_TAG}" "${FROM_PATH}"
  else
    echo "--install_options flag only support helm and istioctl options"
    exit 1
  fi
}

istioUpgradeOptions(){
  if [[ "${INSTALL_OPTIONS}" == "helm" ]];then
    installIstioAtVersionUsingHelm "${TO_HUB}" "${TO_TAG}" "${TO_PATH}"
  elif [[ "${INSTALL_OPTIONS}" == "istioctl" ]];then
    upgradeIstioAtVersionUsingIstioctl "${TO_HUB}" "${TO_TAG}" "${TO_PATH}"
  else
    echo "--install_options flag only support helm and istioctl options"
    exit 1
  fi
}

installTest() {
   writeMsg "Installing test deployments"
   kubectl apply -n "${TEST_NAMESPACE}" -f "${TMP_DIR}/gateway.yaml" || die "kubectl apply gateway.yaml failed"
   # We don't want to auto-inject into ${ISTIO_NAMESPACE}, so echosrv and load client must be in different namespace.
   kubectl apply -n "${TEST_NAMESPACE}" -f "${TMP_DIR}/fortio.yaml" || die "kubectl apply fortio.yaml failed"
   sleep 10
}

# Sends traffic from internal pod (Fortio load command) to Fortio echosrv.
# Since this may block for some time due to restarts, it should be run in the background. Use waitForJob to check for
# completion.
_sendInternalRequestTraffic() {
    local job_name=cli-fortio
    deleteWithWait job "${job_name}" "${TEST_NAMESPACE}"
    start_time=${SECONDS}
    withRetries 10 60 kubectl apply -n "${TEST_NAMESPACE}" -f "${TMP_DIR}/fortio-cli.yaml"
    waitForJob "${job_name}" "${TEST_NAMESPACE}"
    # Any timeouts typically occur in the first 20s
    if (( SECONDS - start_time < 100 )); then
        echo "${job_name} failed"
        return 1
    fi
}

sendInternalRequestTraffic() {
    writeMsg "Sending internal traffic"
    withRetries 10 0 _sendInternalRequestTraffic
}

# Runs traffic from external fortio client, with retries.
runFortioLoadCommand() {
    withRetries 10 10  echo_and_run fortio load -c 32 -t "${TRAFFIC_RUNTIME_SEC}"s -qps 10 -timeout 30s\
        -H "Host:echosrv.test.svc.cluster.local" "http://${1}/echo?size=200" &> "${LOCAL_FORTIO_LOG}"
    echo "done" >> "${EXTERNAL_FORTIO_DONE_FILE}"
}

waitForExternalRequestTraffic() {
    echo "Waiting for external traffic to complete"
    while [[ ! -f "${EXTERNAL_FORTIO_DONE_FILE}" ]]; do
        sleep 10
    done
}

# Sends external traffic from machine test is running on to Fortio echosrv through external IP and ingress gateway LB.
sendExternalRequestTraffic() {
    writeMsg "Sending external traffic"
    runFortioLoadCommand "${1}"
}

resetConfigMap() {
    deleteWithWait ConfigMap "${1}" "${ISTIO_NAMESPACE}"
    kubectl create -n "${ISTIO_NAMESPACE}" -f "${2}"
}

_waitForIngress() {
    INGRESS_HOST=$(kubectl -n "${ISTIO_NAMESPACE}" get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    INGRESS_PORT=$(kubectl -n "${ISTIO_NAMESPACE}" get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
    INGRESS_ADDR=${INGRESS_HOST}:${INGRESS_PORT}
    if [[ -z "${INGRESS_HOST}" ]]; then return 1; fi
}

waitForIngress() {
    echo "Waiting for ingress-gateway addr..."
    withRetriesMaxTime 300 10 _waitForIngress
    echo "Got ingress-gateway addr: ${INGRESS_ADDR}"
}

_checkEchosrv() {
    resp=$( curl -o /dev/null -s -w "%{http_code}\\n" -HHost:echosrv.${TEST_NAMESPACE}.svc.cluster.local "http://${INGRESS_ADDR}/echo" || echo $? )
    if [[ "${resp}" = *"200"* ]]; then
        echo "Got correct response from echosrv."
        return 0
    fi
    echo "Got bad echosrv response: ${resp}"
    return 1
}

checkEchosrv() {
    writeMsg "Checking echosrv..."
    withRetriesMaxTime 300 10 _checkEchosrv
}

copy_test_files

# create cluster admin role binding
user="cluster-admin"
if [[ $CLOUD == "GKE" ]];then
  user="$(gcloud config get-value core/account)"
fi
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="${user}" || echo "clusterrolebinding already created."

resetCluster "${TO_PATH}/bin/istioctl"

istioInstallOptions
waitForIngress
waitForPodsReady "${ISTIO_NAMESPACE}"

if [[ "${TEST_SCENARIO}" == "upgrade-downgrade" ]];then
  # Make a copy of the "from" sidecar injector ConfigMap so that we can restore the sidecar independently later.
  echo_and_run kubectl get ConfigMap -n "${ISTIO_NAMESPACE}" istio-sidecar-injector -o yaml > ${TMP_DIR}/sidecar-injector-configmap.yaml
fi

installTest
waitForPodsReady "${TEST_NAMESPACE}"
checkEchosrv

# Run internal traffic in the background since we may have to relaunch it if the job fails.
sendInternalRequestTraffic &
sendExternalRequestTraffic "${INGRESS_ADDR}" &
# Let traffic clients establish all connections. There's some small startup delay, this covers it.
echo "Waiting for traffic to settle..."
sleep 20

if [[ "${TEST_SCENARIO}" == "upgrade-downgrade" || "${TEST_SCENARIO}" == "upgrade" || "${TEST_SCENARIO}" == "downgrade" ]];then
  istioUpgradeOptions
  waitForPodsReady "${ISTIO_NAMESPACE}"
  # In principle it should be possible to restart data plane immediately, but being conservative here.
  sleep 60

  restartDataPlane echosrv-deployment-v1 "${TEST_NAMESPACE}"
  restartDataPlane echosrv-deployment-v2 "${TEST_NAMESPACE}"
fi

if [[ "${TEST_SCENARIO}" == "upgrade-downgrade" ]];then
  # Now do a rollback. In a rollback, we update the data plane first.
  writeMsg "Starting rollback - first, rolling back data plane to ${FROM_PATH}"
  resetConfigMap istio-sidecar-injector "${TMP_DIR}"/sidecar-injector-configmap.yaml
  
  # echosrv-deployment-v2 is for mTLS traffic
  restartDataPlane echosrv-deployment-v1 "${TEST_NAMESPACE}"
  restartDataPlane echosrv-deployment-v2 "${TEST_NAMESPACE}"

  istioInstallOptions
  waitForPodsReady "${ISTIO_NAMESPACE}"
fi

echo "Test ran for ${SECONDS} seconds."
if (( SECONDS > TRAFFIC_RUNTIME_SEC )); then
    echo "WARNING: test duration was ${SECONDS} but traffic only ran for ${TRAFFIC_RUNTIME_SEC}"
fi

cli_pod_name=$(kubectl -n "${TEST_NAMESPACE}" get pods -lapp=cli-fortio -o jsonpath='{.items[0].metadata.name}')
echo "Traffic client pod is ${cli_pod_name}, waiting for traffic to complete..."
waitForJob cli-fortio "${TEST_NAMESPACE}"
kubectl logs -f -n "${TEST_NAMESPACE}" -c echosrv "${cli_pod_name}" &> "${POD_FORTIO_LOG}" || echo "Could not find ${cli_pod_name}"
waitForExternalRequestTraffic

local_log_str=$(grep "Code 200" "${LOCAL_FORTIO_LOG}")
pod_log_str=$(grep "Code 200"  "${POD_FORTIO_LOG}")

# First dump local log
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

# Then dump pod log
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
    exit 1
fi

echo "SUCCESS"

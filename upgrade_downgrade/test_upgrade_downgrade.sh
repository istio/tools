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

set -x
set -o pipefail

WD=$(dirname "$0")
WD=$(cd "$WD" || exit; pwd)
ROOT=$(dirname "$WD")

command -v fortio >/dev/null 2>&1 || { echo >&2 "fortio must be installed, aborting."; exit 1; }

function usage() {
  echo "Usage:"
  echo "  ./test_upgrade_downgrade.sh [OPTIONS]"
  echo
  echo "  from_hub          hub of release to upgrade from (required)."
  echo "  from_tag          tag of release to upgrade from (required)."
  echo "  from_path         path to release dir to upgrade from (required)."
  echo "  to_hub            hub of release to upgrade to (required)."
  echo "  to_tag            tag of release to upgrade to (required)."
  echo "  to_path           path to release to upgrade to (required)."
  echo "  auth_enable       enable mtls."
  echo "  skip_cleanup      leave install intact after test completes."
  echo "  namespace         namespace to install istio control plane in (default istio-system)."
  echo "  cloud             cloud provider name (required)"
  echo
  echo "  e.g. ./test_upgrade_downgrade.sh \"
  echo "        --from_hub=gcr.io/istio-testing --from_tag=d639408fd --from_path=/tmp/release-d639408fd \"
  echo "        --to_hub=gcr.io/istio-release --to_tag=1.0.2 --to_path=/tmp/istio-1.0.2 --cloud=GKE"
  echo
  exit 1
}

ISTIO_NAMESPACE="istio-system"

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

if [[ -z "${FROM_HUB}" || -z "${FROM_TAG}" || -z "${FROM_PATH}" || -z "${TO_HUB}" || -z "${TO_TAG}" || -z "${TO_PATH}" ]]; then
  echo "Error: from_hub, from_tag, from_path, to_hub, to_tag, to_path must all be set."
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
export TRAFFIC_RUNTIME_SEC=500

# Used to signal that background external process is done.
export EXTERNAL_FORTIO_DONE_FILE=${TMP_DIR}/fortio_done_file

# shellcheck disable=SC1090
source "${ROOT}/upgrade_downgrade/common.sh"
# shellcheck disable=SC1090
source "${ROOT}/upgrade_downgrade/fortio_utils.sh"

function install_istio_at_version_using_istioctl(){
  write_msg "istioctl install istio using version ${2} from ${3}."
  istioctl_path="${3}"/bin
  find "${istioctl_path}" -maxdepth 1 -type f
  "${istioctl_path}"/istioctl install --skip-confirmation
}

function upgrade_istio_at_version_using_istioctl(){
  write_msg "istioctl upgrade istio using version ${2} from ${3}."
  istioctl_path="${3}"/bin
  "${istioctl_path}"/istioctl upgrade --skip-confirmation --charts "${3}"/manifests

}

function istio_install_options() {
  install_istio_at_version_using_istioctl "${FROM_HUB}" "${FROM_TAG}" "${FROM_PATH}"
}

function istio_upgrade_options(){
  upgrade_istio_at_version_using_istioctl "${TO_HUB}" "${TO_TAG}" "${TO_PATH}"
}

function install_test() {
  write_msg "Installing test deployments"
  kubectl apply -n "${TEST_NAMESPACE}" -f "${TMP_DIR}/gateway.yaml" || die "kubectl apply gateway.yaml failed"
  kubectl apply -n "${TEST_NAMESPACE}" -f "${TMP_DIR}/fortio.yaml" || die "kubectl apply fortio.yaml failed"
  sleep 10
}

# Sends traffic from internal pod (Fortio load command) to Fortio echosrv.
# Since this may block for some time due to restarts, it should be run in the background.
function _send_internal_request_traffic() {
  local job_name=cli-fortio
  delete_with_wait job "${job_name}" "${TEST_NAMESPACE}"
  start_time=${SECONDS}
  with_retries 10 60 kubectl apply -n "${TEST_NAMESPACE}" -f "${TMP_DIR}/fortio-cli.yaml"
  kubectl wait --for=condition=complete --timeout=12m "${job_name}" -n "${TEST_NAMESPACE}"
  # Any timeouts typically occur in the first 20s
  if (( SECONDS - start_time < 100 )); then
    echo "${job_name} failed"
    return 1
  fi
}

function send_internal_request_traffic() {
  write_msg "Sending internal traffic"
  with_retries 10 0 _send_internal_request_traffic
}

function reset_config_map() {
  delete_with_wait ConfigMap "${1}" "${ISTIO_NAMESPACE}"
  kubectl create -n "${ISTIO_NAMESPACE}" -f "${2}"
}

function _check_echosrv() {
  resp=$( curl -o /dev/null -s -w "%{http_code}\\n" -HHost:echosrv.${TEST_NAMESPACE}.svc.cluster.local "http://${INGRESS_ADDR}/echo" || echo $? )
  if [[ "${resp}" = *"200"* ]]; then
    echo "Got correct response from echosrv."
    return 0
  fi
  echo "Got bad echosrv response: ${resp}"
  return 1
}

function check_echosrv() {
  write_msg "Checking echosrv..."
  with_retries_max_time 300 10 _check_echosrv
}

copy_test_files

# create cluster admin role binding
user="cluster-admin"
if [[ $CLOUD == "GKE" ]];then
  user="$(gcloud config get-value core/account)"
fi
kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="${user}" || echo "clusterrolebinding already created."

reset_cluster "${TO_PATH}/bin/istioctl"

istio_install_options || die "istio installation failed"
wait_for_ingress
kubectl wait --for=condition=ready --timeout=10m pod --all -n "${ISTIO_NAMESPACE}"

if [[ "${TEST_SCENARIO}" == "upgrade-downgrade" ]];then
  # Make a copy of the "from" sidecar injector ConfigMap so that we can restore the sidecar independently later.
  echo_and_run kubectl get ConfigMap -n "${ISTIO_NAMESPACE}" istio-sidecar-injector -o yaml > ${TMP_DIR}/sidecar-injector-configmap.yaml
fi

install_test
kubectl wait --for=condition=ready --timeout=10m pod --all -n "${TEST_NAMESPACE}"
check_echosrv

# Run internal traffic in the background since we may have to relaunch it if the job fails.
send_internal_request_traffic &

write_msg "sending external traffic. Ingress=${INGRESS_ADDR}"
send_external_request_traffic "http://${INGRESS_ADDR}/echo?size=200" -H "Host:echosrv.test.svc.cluster.local" &
# Let traffic clients establish all connections. There's some small startup delay, this covers it.
echo "Waiting for traffic to settle..."
sleep 20

if [[ "${TEST_SCENARIO}" == "upgrade-downgrade" || "${TEST_SCENARIO}" == "upgrade" || "${TEST_SCENARIO}" == "downgrade" ]];then
  # We should have failed the job if it fails. However, if there is an unreleased
  # version, then the command fails. One way to detect this is to check if 'master'
  # is passed from run_upgrade_downgrade.sh. Currently, we pass the actual tag.
  istio_upgrade_options || [[ -z "${UNRELEASED_VERSION_INVOLVED}" ]] && die "upgrade/downgrade failed"
  kubectl wait --for=condition=ready --timeout=10m pod --all -n "${ISTIO_NAMESPACE}"
  # In principle it should be possible to restart data plane immediately, but being conservative here.
  sleep 60

  restart_data_plane echosrv-deployment-v1 "${TEST_NAMESPACE}"
  restart_data_plane echosrv-deployment-v2 "${TEST_NAMESPACE}"
fi

if [[ "${TEST_SCENARIO}" == "upgrade-downgrade" ]];then
  # Now do a rollback. In a rollback, we update the data plane first.
  write_msg "Starting rollback - first, rolling back data plane to ${FROM_PATH}"
  reset_config_map istio-sidecar-injector "${TMP_DIR}"/sidecar-injector-configmap.yaml
  
  # echosrv-deployment-v2 is for mTLS traffic
  restart_data_plane echosrv-deployment-v1 "${TEST_NAMESPACE}"
  restart_data_plane echosrv-deployment-v2 "${TEST_NAMESPACE}"

  istio_install_options || echo "istio installation failed"
  kubectl wait --for=condition=ready --timeout=10m pod --all -n "${ISTIO_NAMESPACE}"
fi

echo "Test ran for ${SECONDS} seconds."
if (( SECONDS > TRAFFIC_RUNTIME_SEC )); then
  echo "WARNING: test duration was ${SECONDS} but traffic only ran for ${TRAFFIC_RUNTIME_SEC}"
fi

cli_pod_name=$(kubectl -n "${TEST_NAMESPACE}" get pods -lapp=cli-fortio -o jsonpath='{.items[0].metadata.name}')
echo "Traffic client pod is ${cli_pod_name}, waiting for traffic to complete..."
kubectl wait --for=condition=complete --timeout=30m job/cli-fortio -n "${TEST_NAMESPACE}"
kubectl logs -f -n "${TEST_NAMESPACE}" -c echosrv "${cli_pod_name}" &> "${POD_FORTIO_LOG}" || echo "Could not find ${cli_pod_name}"
wait_for_external_request_traffic

MAX_503_PCT_FOR_PASS="15"
MAX_CONNECTION_ERR_FOR_PASS="30"

if ! analyze_fortio_logs "${POD_FORTIO_LOG}" "${MAX_503_PCT_FOR_PASS}" "${MAX_CONNECTION_ERR_FOR_PASS}"; then
  failed=true
elif ! analyze_fortio_logs "${LOCAL_FORTIO_LOG}" "${MAX_503_PCT_FOR_PASS}" "${MAX_CONNECTION_ERR_FOR_PASS}"; then
  failed=true
fi

if [[ -n "${failed}" ]]; then
  exit 1
fi

echo "SUCCESS"

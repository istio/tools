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
        # shellcheck disable=SC2034
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
if [[ "${TEST_SCENARIO}" == "multicluster-dual-control-plane-upgrade" ]];then
  echo "The current test scenario is ${TEST_SCENARIO}."
else
  echo "Invalid scenario: ${TEST_SCENARIO}"
  echo "supported: multicluster-dual-control-plane-upgrade"
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

TMP_DIR=/tmp/istio_multicluster_upgrade_test
TEST_NAMESPACE="test"

TO_REVISION=$(echo "${TO_TAG}" | tr '.' '-' | cut -c -20)
FROM_REVISION=$(echo "${FROM_TAG}" | tr '.' '-' | cut -c -20)

# KUBECONFIG is supplied from multicluster setup
if [[ -z "${KUBECONFIG}" ]]; then
  echo "KUBECONFIG environment variable is not set"
  exit 1
fi
echo "${KUBECONFIG}"

write_msg "Kubernetes context"
echo "$KUBECONFIG"
kubectl config get-contexts
kubectl config view

function establish_root_of_trust() {
  local cert_dir="${TMP_DIR}/certs"
  local cert_gen_dir="${cert_dir}/generator"
  mkdir -p "$cert_gen_dir"

  wget -O "$cert_gen_dir/Makefile.k8s.mk" "https://raw.githubusercontent.com/istio/istio/master/tools/certs/Makefile.k8s.mk"
  wget -O "$cert_gen_dir/Makefile.selfsigned.mk" "https://raw.githubusercontent.com/istio/istio/master/tools/certs/Makefile.selfsigned.mk"
  wget -O "$cert_gen_dir/common.mk" "https://raw.githubusercontent.com/istio/istio/master/tools/certs/common.mk"

  # shellcheck disable=SC2164
  pushd "${cert_dir}"
  make -f "${cert_gen_dir}/Makefile.selfsigned.mk" root-ca
  # shellcheck disable=SC2164
  popd

  for cluster in ${CLUSTERS//:/ }; do
    # shellcheck disable=SC2164
    pushd "${cert_dir}"
    make -f "${cert_gen_dir}/Makefile.selfsigned.mk" "${cluster}-cacerts"
    # shellcheck disable=SC2164
    popd

    kubectl config use-context "kind-${cluster}"
    kubectl create secret generic cacerts -n "${ISTIO_NAMESPACE}" \
      --from-file="${cert_dir}/${cluster}/ca-cert.pem" \
      --from-file="${cert_dir}/${cluster}/ca-key.pem" \
      --from-file="${cert_dir}/${cluster}/root-cert.pem" \
      --from-file="${cert_dir}/${cluster}/cert-chain.pem"
  done
}

function install_with_iop() {
  local istioctl_path="${1}"
  local iop_path="${2}"
  local revision="${3}"
  local context="${4}"
  "${istioctl_path}" --context="${context}" install -y -f "${iop_path}" --revision="${revision}" \
    || die "installation failed (path=$iop_path, revision=$revision, context=$context)"
}

copy_test_files

for cluster in ${CLUSTERS//:/ }; do
  write_msg "Reset cluster: $cluster"
  kubectl config use-context "kind-$cluster"
  reset_cluster "${TO_ISTIOCTL}"
done

write_msg "Establish common root of trust between clusters"
establish_root_of_trust

# shellcheck disable=SC2206
CLUSTER_NAMES=(${CLUSTERS//:/ })
CTX_CLUSTER1="kind-${CLUSTER_NAMES[0]}"
CTX_CLUSTER2="kind-${CLUSTER_NAMES[1]}"
MULTICLUSTER_IOP_PATH="${TMP_DIR}/multicluster/multi-primary"

write_msg "Add topology label to ${ISTIO_NAMESPACE} (multi-primary, multi-network topology)"
kubectl --context="${CTX_CLUSTER1}" label namespace "${ISTIO_NAMESPACE}" topology.istio.io/network=network1
kubectl --context="${CTX_CLUSTER2}" label namespace "${ISTIO_NAMESPACE}" topology.istio.io/network=network2

write_msg "Install Istiod ($FROM_REVISION) on both clusters"
install_with_iop "${FROM_ISTIOCTL}" "${MULTICLUSTER_IOP_PATH}/cluster1.yaml" "${FROM_REVISION}" "${CTX_CLUSTER1}"
install_with_iop "${FROM_ISTIOCTL}" "${MULTICLUSTER_IOP_PATH}/cluster2.yaml" "${FROM_REVISION}" "${CTX_CLUSTER2}"

write_msg "Install ingress-gateway ($FROM_REVISION) on both clusters"
install_with_iop "${FROM_ISTIOCTL}" "${MULTICLUSTER_IOP_PATH}/ingress1.yaml" "${FROM_REVISION}" "${CTX_CLUSTER1}"
install_with_iop "${FROM_ISTIOCTL}" "${MULTICLUSTER_IOP_PATH}/ingress2.yaml" "${FROM_REVISION}" "${CTX_CLUSTER2}"

write_msg "Install eastwest-gateway ($FROM_REVISION) on both clusters"
install_with_iop "${FROM_ISTIOCTL}" "${MULTICLUSTER_IOP_PATH}/eastwest1.yaml" "${FROM_REVISION}" "${CTX_CLUSTER1}"
install_with_iop "${FROM_ISTIOCTL}" "${MULTICLUSTER_IOP_PATH}/eastwest2.yaml" "${FROM_REVISION}" "${CTX_CLUSTER2}"

kubectl --context="${CTX_CLUSTER1}" wait --for=condition=ready pods --all -n "${ISTIO_NAMESPACE}" --timeout=15m
kubectl --context="${CTX_CLUSTER2}" wait --for=condition=ready pods --all -n "${ISTIO_NAMESPACE}" --timeout=15m

write_msg "Expose services to remote clusters"
kubectl --context="${CTX_CLUSTER1}" apply -n "${ISTIO_NAMESPACE}" -f "${MULTICLUSTER_IOP_PATH}/../cross-network-gateway.yaml"
"${FROM_ISTIOCTL}" x wait --for=distribution gateway cross-network-gateway.${ISTIO_NAMESPACE} --context="${CTX_CLUSTER1}"

kubectl --context="${CTX_CLUSTER2}" apply -n "${ISTIO_NAMESPACE}" -f "${MULTICLUSTER_IOP_PATH}/../cross-network-gateway.yaml"
"${FROM_ISTIOCTL}" x wait --for=distribution gateway cross-network-gateway.${ISTIO_NAMESPACE} --context="${CTX_CLUSTER2}"

write_msg "Expose api-server for Istiod instances in remote clusters"
"${FROM_ISTIOCTL}" x create-remote-secret --context="${CTX_CLUSTER1}" \
    --name="kind-${CLUSTER_NAMES[0]}" | kubectl apply -f - --context="${CTX_CLUSTER2}"
"${FROM_ISTIOCTL}" x create-remote-secret --context="${CTX_CLUSTER2}" \
    --name="kind-${CLUSTER_NAMES[1]}" | kubectl apply -f - --context="${CTX_CLUSTER1}"

write_msg "Install application in ${TEST_NAMESPACE} in both clusters"
kubectl --context="${CTX_CLUSTER1}" label namespace "${TEST_NAMESPACE}" istio.io/rev="${FROM_REVISION}"
kubectl --context="${CTX_CLUSTER2}" label namespace "${TEST_NAMESPACE}" istio.io/rev="${FROM_REVISION}"

HELLOWORLD_URL="https://raw.githubusercontent.com/istio/istio/master/samples/helloworld/helloworld.yaml"
kubectl apply -f "${HELLOWORLD_URL}" --context="${CTX_CLUSTER1}" -l service=helloworld -n "${TEST_NAMESPACE}"
kubectl apply -f "${HELLOWORLD_URL}" --context="${CTX_CLUSTER2}" -l service=helloworld -n "${TEST_NAMESPACE}"

kubectl apply -f "${HELLOWORLD_URL}" --context="${CTX_CLUSTER1}" -l version=v1 -n "${TEST_NAMESPACE}"
kubectl wait --for=condition=ready pods --all -n "${TEST_NAMESPACE}" --context="${CTX_CLUSTER1}" --timeout=12m

kubectl apply -f "${HELLOWORLD_URL}" --context="${CTX_CLUSTER2}" -l version=v2 -n "${TEST_NAMESPACE}"
kubectl wait --for=condition=ready pods --all -n "${TEST_NAMESPACE}" --context="${CTX_CLUSTER2}" --timeout=12m

write_msg "Send external traffic through fortio on ingressgateway for both clusters"

# First, find the address of the ingress gateway for both clusters
kubectl config use-context "${CTX_CLUSTER1}"
wait_for_ingress
# shellcheck disable=SC2153
INGRESS_ADDR1="${INGRESS_ADDR}"

kubectl config use-context "${CTX_CLUSTER2}"
wait_for_ingress
INGRESS_ADDR2="${INGRESS_ADDR}"

FORTIO_LOG1="${TMP_DIR}/fortio_local_1.log"
FORTIO_LOG2="${TMP_DIR}/fortio_local_2.log"

export TRAFFIC_RUNTIME_SEC
export LOCAL_FORTIO_LOG
export EXTERNAL_FORTIO_DONE_FILE

TRAFFIC_RUNTIME_SEC=600
LOCAL_FORTIO_LOG="${FORTIO_LOG1}"
EXTERNAL_FORTIO_DONE_FILE="${TMP_DIR}/fortio_1_done"
send_external_request_traffic "http://${INGRESS_ADDR1}/hello" -H "Host:helloworld.test.svc.cluster.local" &

LOCAL_FORTIO_LOG="${FORTIO_LOG2}"
EXTERNAL_FORTIO_DONE_FILE="${TMP_DIR}/fortio_2_done"
send_external_request_traffic "http://${INGRESS_ADDR2}/hello" -H "Host:helloworld.test.svc.cluster.local" &

# Next, send traffic through those gateways

#write_msg "Verify load balancing between clusters"
#
#write_msg "Install Istiod ($TO_REVISION) on cluster1"
#
#write_msg "Install ingress-gateway ($TO_REVISION) on cluster1"
#
#write_msg "Install eastwest-gateway ($TO_REVISION) on cluster1"
#
#write_msg "Restart data plane in cluster1 to point to $TO_REVISION"
#
#write_msg "Install Istiod ($TO_REVISION) on cluster2"
#
#write_msg "Install ingress-gateway ($TO_REVISION) on cluster2"
#
#write_msg "Install eastwest-gateway ($TO_REVISION) on cluster2"
#
#write_msg "Restart data plane in cluster2 to point to $TO_REVISION"

write_msg "Waiting for Fortio traffic to complete"
EXTERNAL_FORTIO_DONE_FILE="${TMP_DIR}/fortio_1_done" wait_for_external_request_traffic
EXTERNAL_FORTIO_DONE_FILE="${TMP_DIR}/fortio_2_done" wait_for_external_request_traffic

write_msg "Analyzing fortio logs from both clusters and external gateways"
MAX_503_PCT_FOR_PASS="5"
MAX_CONNECTION_ERR_FOR_PASS="30"

if ! analyze_fortio_logs "${FORTIO_LOG1}" "${MAX_503_PCT_FOR_PASS}" "${MAX_CONNECTION_ERR_FOR_PASS}"; then
  failed=1
elif ! analyze_fortio_logs "${FORTIO_LOG2}" "${MAX_503_PCT_FOR_PASS}" "${MAX_CONNECTION_ERR_FOR_PASS}"; then
  failed=1
fi

if [[ -n "${failed}" ]]; then
  echo "test failed"
  bash
  exit 1
fi

echo "SUCCESS"

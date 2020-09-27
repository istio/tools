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

set -e
set -o pipefail
set -x

# TODO: Remove hardcoded versions and pass it from run_upgrade_downgrade_test
# as command-line arguments like test_upgrade_downgrade.sh
FROM_VERSION="istio-1.7-alpha.a8c04f92e236676977b3ed58132437d9fafc9aed"
TO_VERSION="istio-1.8-alpha.f0f221ae1024fb2d4a93b355546f094376fb2e47"
ISTIOPATH="$GOPATH/src/istio.io"

FROM_DIR="${ISTIOPATH}/tools/upgrade_downgrade/from_dir.g9xSYb/${FROM_VERSION}"
TO_DIR="${ISTIOPATH}/tools/upgrade_downgrade/to_dir.MrCCp4/${TO_VERSION}"
TMP_DIR="${ISTIOPATH}/tools/upgrade_downgrade/templates"
POD_FORTIO_LOG="${TMP_DIR}/fortio.log"

FROM_ISTIOCTL="${FROM_DIR}/bin/istioctl"
TO_ISTIOCTL="${TO_DIR}/bin/istioctl"

# Maximum % of 503 response that cannot exceed
MAX_503_PCT_FOR_PASS="15"
# Maximum % of connection refused that cannot exceed
# Set it to high value so it fails for explicit sidecar issues
MAX_CONNECTION_ERR_FOR_PASS="30"
SERVICE_UNAVAILABLE_CODE="503"
CONNECTION_ERROR_CODE="-1"


# Install Istio 1.7 minimal profile (only istiod)
echo "Installing Istio ${FROM_VERSION}"
${FROM_ISTIOCTL} install -y --set profile=minimal
kubectl wait --all --for=condition=Ready pods -n istio-system --timeout=5m

# Deploy sample application
kubectl create namespace test
kubectl label namespace test istio-injection=enabled

echo "Deploying Echo v1 and v2"
kubectl apply -f "${TMP_DIR}/fortio.yaml" -n test
kubectl wait --all --for=condition=Ready pods -n test --timeout=5m

echo "Generate internal traffic for echo v1 and v2"
kubectl apply -f "${TMP_DIR}/fortio-cli.yaml" -n test

# Install Istio 1.8 minimal profile with canary revision
echo "Installng Istio-canary ${TO_VERSION}"
${TO_ISTIOCTL} install -y --set profile=minimal --set revision=canary
kubectl wait --all --for=condition=Ready pods -n istio-system --timeout=5m

# Relabel namespace before restarting each service
echo "Relabel namespace to inject canary-release proxy"
kubectl label namespace test istio-injection-
kubectl label namespace test istio.io/rev=canary

function rolloutDeployment() {
  local ns="$1"
  local name="$2"
  local max_attempts=${3:-30}
  local cur_attempt=1

  local total_replicas=$(kubectl get deployment "${name}" -n "${ns}" -o jsonpath='{.spec.replicas}')
  kubectl rollout restart deployment "${name}" -n "${ns}"

  while (( $cur_attempt <= $max_attempts )); do
    local ready=$(kubectl get deployment "${name}" -n "${ns}" -o jsonpath='{.status.readyReplicas}')
    local updated=$(kubectl get deployment "${name}" -n "${ns}" -o jsonpath='{.status.updatedReplicas}')
    local available=$(kubectl get deployment "${name}" -n "${ns}" -o jsonpath='{.status.availableReplicas}')
    
    echo "attempt: ${cur_attempt}/${max_attempts} ==> ready=${ready}, updated=${updated}, available=${available}"
    if ((updated == total_replicas && available == total_replicas)); then
      echo "rollout complete"
      return 0
    fi
    sleep 10
    cur_attempt=$(( $cur_attempt+1 ))
  done

  echo "timed out waiting for full rollout"
  return 1
}

function verifyIstiod() {
  local ns="$1"
  local app="$2"
  local version="$3"
  local istioctl_path="$4"
  local expected="$5"

  local cur_attempt=1
  local max_attempts=5
  
  while (( $cur_attempt <= $max_attempts )); do
    local mismatch=0
    echo "attempt ${cur_attempt}/${max_attempts}"
    for pod in $(kubectl get pod -lapp="$app" -lversion="$version" -n "$ns" -o name); do
      local istiod=$(${istioctl_path} proxy-config endpoint "$pod.test" --cluster xds-grpc -o json | jq -r '.[].hostStatuses[].hostname')
      echo "  $pod ==> ${istiod}"
      if [[ "$istiod" != *"$expected"* ]]; then
        mismatch=$(( $mismatch+1 ))
      fi
    done

    if (($mismatch == 0)); then
      return 0
    fi
    sleep 20
    cur_attempt=$(($cur_attempt+1))
    echo '=========='
  done

  echo "timeout out while trying to match istiod=$expected"
  return 1
}

function waitForJob() {
  local ns="$1"
  local job="$2"
  until kubectl get jobs -n "${ns}" "${job}" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' | grep True ; do
    sleep 5;
  done
  echo "job ${job}.${ns} done...."
}

rolloutDeployment test echosrv-deployment-v1
verifyIstiod "test" "echosrv-deployment-v1" "v1" "${TO_ISTIOCTL}" "istiod-canary.istio-system.svc"

rolloutDeployment test echosrv-deployment-v2
verifyIstiod "test" "echosrv-deployment-v2" "v2" "${TO_ISTIOCTL}" "istiod-canary.istio-system.svc"

# List all objects in istio-system namespace
# It should have istiod and istiod-canary
${FROM_ISTIOCTL} experimental uninstall --filename "${FROM_DIR}/manifests/profiles/minimal.yaml"
kubectl get all -n istio-system

cli_pod_name=$(kubectl -n test get pods -lapp=cli-fortio -o jsonpath='{.items[0].metadata.name}')
waitForJob test cli-fortio
kubectl logs -f -n test -c echosrv "${cli_pod_name}" &> "${POD_FORTIO_LOG}" || echo "Could not find ${cli_pod_name}"
pod_log_str=$(grep "Code 200"  "${POD_FORTIO_LOG}")

cat ${POD_FORTIO_LOG}

# Return 1 if the specific error code percentage exceed corresponding threshold
errorPercentBelow() {
  local LOG=${1}
  local ERR_CODE=${2}
  local LIMIT=${3}
  local s
  s=$(grep "Code ${ERR_CODE}" "${LOG}")
  local regex="Code ${ERR_CODE} : [0-9]+ \\(([0-9]+)\\.[0-9]+ %\\)"
  if [[ ${s} =~ ${regex} ]]; then
    local pctErr="${BASH_REMATCH[1]}"
    if (( pctErr > LIMIT )); then
      return 1
    fi
    echo "Errors percentage is within threshold"
  fi
  return 0
}

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

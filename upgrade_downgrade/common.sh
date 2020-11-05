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

die() {
  echo "$*" 1>&2 ; exit 1;
}

echo_and_run() { echo "# RUNNING $*" ; "$@" ; }
echo_and_run_quiet() { echo "# RUNNING(quiet) $*" ; "$@" > /dev/null 2>&1 ; }
echo_and_run_or_die() { echo "# RUNNING $*" ; "$@" || die "failed!" ; }

writeMsg() {
  printf "\\n\\n****************\\n\\n%s\\n\\n****************\\n\\n" "${1}"
}

# withRetries retries the given command ${1} times with ${2} sleep between retries
# e.g. withRetries 10 60 myFunc param1 param2
#   runs "myFunc param1 param2" up to 10 times with 60 sec sleep in between.
withRetries() {
  local max_retries=${1}
  local sleep_sec=${2}
  local n=0
  shift
  shift
  while (( n < max_retries )); do
    echo "RUNNING $*" ; "${@}" && break
    echo "Failed, sleeping ${sleep_sec} seconds and retrying..."
    ((n++))
    sleep "${sleep_sec}"
  done

  if (( n == max_retries )); then die "$* failed after retrying ${max_retries} times."; fi
  echo "Succeeded."
}

# withRetriesMaxTime retries the given command repeatedly with ${2} sleep between retries until ${1} seconds have elapsed.
# e.g. withRetries 300 60 myFunc param1 param2
#   runs "myFunc param1 param2" for up 300 seconds with 60 sec sleep in between.
withRetriesMaxTime() {
  local total_time_max=${1}
  local sleep_sec=${2}
  local start_time=${SECONDS}
  shift
  shift
  while (( SECONDS - start_time <  total_time_max )); do
    echo "RUNNING $*" ; "${@}" && break
    echo "Failed, sleeping ${sleep_sec} seconds and retrying..."
    sleep "${sleep_sec}"
  done

  if (( SECONDS - start_time >=  total_time_max )); then die "$* failed after retrying for ${total_time_max} seconds."; fi
  echo "Succeeded."
}

# checkIfDeleted checks if a resource has been deleted, returns 1 if it has not.
# e.g. checkIfDeleted ConfigMap my-config-map istio-system
#   OR checkIfDeleted namespace istio-system
checkIfDeleted() {
  local resp
  if [[ -n "${3}" ]]; then
      resp=$( kubectl get "${1}" -n "${3}" "${2}" 2>&1 )
  else
      resp=$( kubectl get "${1}" "${2}" 2>&1 )
  fi
  if [[ "${resp}" == *"Error from server (NotFound)"* ]]; then
      return 0
  fi
  echo "Response from server for kubectl get: "
  echo "${resp}"
  return 1
}

deleteWithWait() {
  # Don't complain if resource is already deleted.
  echo_and_run_quiet kubectl delete "${1}" -n "${3}" "${2}"
  withRetries 60 10 checkIfDeleted "${1}" "${2}" "${3}"
}

_waitForPodsReady() {
  pods_str=$(kubectl -n "${1}" get pods | tail -n +2 )
  arr=()
  while read -r line; do
    arr+=("$line")
  done <<< "$pods_str"

  ready="true"
  for line in "${arr[@]}"; do
    if [[ ${line} != *"Running"* && ${line} != *"Completed"* ]]; then
      ready="false"
    fi
  done
  if [[  "${ready}" = "true" ]]; then
    return 0
  fi

  echo "${pods_str}"
  return 1
}

waitForPodsReady() {
  echo "Waiting for pods to be ready in ${1}..."
  withRetriesMaxTime 600 10 _waitForPodsReady "${1}"
  echo "All pods ready."
}

waitForJob() {
  echo "Waiting for job ${1} to complete..."
  local start_time=${SECONDS}
  until kubectl get jobs -n "${2}" "${1}" -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' | grep True ; do
      sleep 1 ;
  done
  run_time=0
  (( run_time = SECONDS - start_time ))
  echo "Job ${1} ran for ${run_time} seconds."
}

restartDataPlane() {
  local name="$1"
  local namespace="$2"
  writeMsg "Restarting deployment ${namespace}/${name}"
  echo_and_run_or_die kubectl rollout restart deployment "${name}" -n "${namespace}"
  echo_and_run_or_die kubectl rollout status deployment "${name}" -n "${namespace}" --timeout=30m  
}

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

# Make a copy of test manifests in case either to/from branch doesn't contain them.
copy_test_files() {
  rm -Rf "${TMP_DIR}"
  mkdir -p "${TMP_DIR}"
  echo "${WD}"
  cp -f "${WD}"/templates/* "${TMP_DIR}"/.
}

resetCluster() {
  echo "Removing Istio CRDs"
  
  # Ideally we should use `istioctl x uninstall --purge -y`
  # But istioctl < 1.7 does not seem to support it. In order
  # not to make things complicated, I'm removing CRDs
  local istioctl=${1}
  "${istioctl}" x uninstall --purge -y
  
  echo "Cleaning cluster by removing namespaces ${ISTIO_NAMESPACE} and ${TEST_NAMESPACE}"
  deleteWithWait namespace "${ISTIO_NAMESPACE}"
  deleteWithWait namespace "${TEST_NAMESPACE}"
  echo "All namespaces deleted. Recreating ${ISTIO_NAMESPACE} and ${TEST_NAMESPACE}"

  echo_and_run_or_die kubectl create namespace "${ISTIO_NAMESPACE}"
  echo_and_run_or_die kubectl create namespace "${TEST_NAMESPACE}"
  echo_and_run_or_die kubectl label namespace "${TEST_NAMESPACE}" istio-injection=enabled
}

function getIstiod() {
  local istioctl_path="${1}"
  local podname="${2}"
  local ns="${3}"
  ${istioctl_path} proxy-config bootstrap "$podname.$ns" | jq -r '.bootstrap.node.metadata.PROXY_CONFIG.discoveryAddress')
}

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
    istiod=$(getIstiod "${istioctl_path}" "${podname}" "${ns}")
    if [[ "$istiod" != *"$expected"* ]]; then
      mismatch=$(( mismatch+1 ))
    fi
  done

  if ((mismatch == 0)); then
    return 0
  fi
  return 1
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

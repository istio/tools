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

function die() {
  echo "$*" 1>&2 ; exit 1;
}

function echo_and_run() { echo "# RUNNING $*" ; "$@" ; }
function echo_and_run_quiet() { echo "# RUNNING(quiet) $*" ; "$@" > /dev/null 2>&1 ; }
function echo_and_run_or_die() { echo "# RUNNING $*" ; "$@" || die "failed!" ; }

function write_msg() {
  printf "\\n\\n****************\\n\\n%s\\n\\n****************\\n\\n" "${1}"
}

# with_retries retries the given command ${1} times with ${2} sleep between retries
# e.g. with_retries 10 60 myFunc param1 param2
#   runs "myFunc param1 param2" up to 10 times with 60 sec sleep in between.
function with_retries() {
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

# with_retries_max_time retries the given command repeatedly with ${2} sleep between retries until ${1} seconds have elapsed.
# e.g. with_retries 300 60 myFunc param1 param2
#   runs "myFunc param1 param2" for up 300 seconds with 60 sec sleep in between.
function with_retries_max_time() {
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

# check_if_deleted checks if a resource has been deleted, returns 1 if it has not.
# e.g. check_if_deleted ConfigMap my-config-map istio-system
#   OR check_if_deleted namespace istio-system
function check_if_deleted() {
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

function delete_with_wait() {
  # Don't complain if resource is already deleted.
  if [[ -n "${3}" ]]; then
    echo_and_run_quiet kubectl delete "${1}" -n "${3}" "${2}"
  else
    # Useful for cluster scoped resources
    echo_and_run_quiet kubectl delete "${1}" "${2}"
  fi
  with_retries 60 10 check_if_deleted "${1}" "${2}" "${3}"
}

function _wait_for_pods_ready() {
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

function wait_for_pods_ready() {
  echo "Waiting for pods to be ready in ${1}..."
  with_retries_max_time 900 10 _wait_for_pods_ready "${1}"
  echo "All pods ready."
}

function restart_data_plane() {
  local name="$1"
  local namespace="$2"
  write_msg "Restarting deployment ${namespace}/${name}"
  echo_and_run_or_die kubectl rollout restart deployment "${name}" -n "${namespace}"
  echo_and_run_or_die kubectl rollout status deployment "${name}" -n "${namespace}" --timeout=30m  
}

# Make a copy of test manifests in case either to/from branch doesn't contain them.
function copy_test_files() {
  rm -Rf "${TMP_DIR}"
  mkdir -p "${TMP_DIR}"
  echo "${WD}"
  cp -f -a "${WD}"/templates/* "${TMP_DIR}"/.
}

function reset_cluster() {
  echo "Removing Istio CRDs"
  
  # Ideally we should use `istioctl x uninstall --purge -y`
  # But istioctl < 1.7 does not seem to support it. In order
  # not to make things complicated, I'm removing CRDs
  local istioctl=${1}
  "${istioctl}" x uninstall --purge -y

  ISTIO_NAMESPACE="${ISTIO_NAMESPACE:-istio-system}"
  TEST_NAMESPACE="${TEST_NAMESPACE:-test}"
  LOADGEN_NAMESPACE="${LOADGEN_NAMESPACE:-loadgen}"
  
  echo "Cleaning cluster by removing namespaces ${ISTIO_NAMESPACE}, ${TEST_NAMESPACE} and ${LOADGEN_NAMESPACE}"
  delete_with_wait namespace "${ISTIO_NAMESPACE}"
  delete_with_wait namespace "${TEST_NAMESPACE}"
  delete_with_wait namespace "${LOADGEN_NAMESPACE}"
  echo "All namespaces deleted. Recreating ${ISTIO_NAMESPACE}, ${TEST_NAMESPACE} and ${LOADGEN_NAMESPACE}"

  echo_and_run_or_die kubectl create namespace "${ISTIO_NAMESPACE}"
  echo_and_run_or_die kubectl create namespace "${TEST_NAMESPACE}"
  echo_and_run_or_die kubectl create namespace "${LOADGEN_NAMESPACE}"
  echo_and_run_or_die kubectl label namespace "${TEST_NAMESPACE}" istio-injection=enabled
}

function _wait_for_ingress() {
    INGRESS_HOST=$(kubectl -n "${ISTIO_NAMESPACE}" get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    INGRESS_PORT=$(kubectl -n "${ISTIO_NAMESPACE}" get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
    INGRESS_ADDR=${INGRESS_HOST}:${INGRESS_PORT}
    if [[ -z "${INGRESS_HOST}" ]]; then return 1; fi
}

function wait_for_ingress() {
    echo "Waiting for ingress-gateway addr..."
    with_retries_max_time 300 10 _wait_for_ingress
    echo "Got ingress-gateway addr: ${INGRESS_ADDR}"
}

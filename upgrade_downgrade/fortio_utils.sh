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

# Return 1 if the specific error code percentage exceed corresponding threshold
function error_percent_below() {
  local LOG=${1}
  local ERR_CODE=${2}
  local LIMIT=${3}
  local s
  local pct
  s=$(grep "Code.*${ERR_CODE}" "${LOG}" | tr 'A-Za-z%:()' ' ' | sed "s/ \+/ /g")
  if [[ -n "${s}" ]]; then
    pct=$(echo "$s" | cut -d' ' -f4)
    if [[ $(python -c "print($pct > $LIMIT)") == *True* ]]; then
      return 1
    fi
    echo "errors encountered ($pct) are within threshold ($LIMIT)"
  fi
}

function run_fortio_load_command() {
  local url="$1"
  shift
  if [[ -z "${TRAFFIC_RUNTIME_SEC}" ]]; then
    echo "TRAFFIC_RUNTIME_SEC is not defined. Setting it to 500s"
    TRAFFIC_RUNTIME_SEC=500
  fi
  if [[ -z "${LOCAL_FORTIO_LOG}" || -z "${EXTERNAL_FORTIO_DONE_FILE}" ]]; then
    echo "fatal: LOCAL_FORTIO_LOG and EXTERNAL_FORTIO_DONE_FILE are not defined"
    exit 1
  fi
  if [[ -z "${url}" ]]; then
    echo "fatal: URL is not specified"
    exit 1
  fi
  fortio load -c 32 -t "${TRAFFIC_RUNTIME_SEC}"s -qps 10 -timeout 30s $@ "${url}" &> "${LOCAL_FORTIO_LOG}"
  echo "done" >> "${EXTERNAL_FORTIO_DONE_FILE}"
}

function wait_for_external_request_traffic() {
  if [[ -z "${EXTERNAL_FORTIO_DONE_FILE}" ]]; then
    echo "fatal: EXTERNAL_FORTIO_DONE_FILE is not defined"
    exit 1
  fi
  echo "Waiting for external traffic to complete"
  local attempt=0
  while [[ ! -f "${EXTERNAL_FORTIO_DONE_FILE}" ]]; do
    echo "attempt ${attempt}"
    attempt=$((attempt+1))
    sleep 10
  done
}

function send_external_request_traffic() {
  local addr="${1}"
  shift
  if [[ -z "${addr}" ]]; then
    echo "fatal: cannot send traffic. INGRESS_ADDR is not set"
    exit 1
  fi
  echo "Sending external traffic"
  run_fortio_load_command "${addr}" "${@}"
}

function analyze_fortio_logs() {
  local fortio_log_file="${1}"
  local status_503_threshold="${2}"
  local conn_error_threshold="${3}"
  if [[ ! -f "${fortio_log_file}" ]]; then
    echo "fortio log file is not specified"
    exit 1
  fi
  if [[ -z "${status_503_threshold}" || -z "${conn_error_threshold}" ]]; then
    echo "status_503_threshold and conn_error_threshold should be passed for fortio log analysis"
    exit 1
  fi
  
  echo "analyzing fortio log file ${fortio_log_file}"
  cat "${fortio_log_file}"
  local code_200_line
  code_200_line=$(grep "Code 200" "${fortio_log_file}")

  if [[ ${code_200_line} != *"Code 200"* ]];then
    echo "=== No Code 200 found in log ==="
    failed=true
  elif ! error_percent_below "${fortio_log_file}" "503" ${status_503_threshold}; then
    echo "=== Code 503 Errors found in traffic exceeded ${status_503_threshold}% threshold ==="
    failed=true
  elif ! error_percent_below "${fortio_log_file}" "-1" ${conn_error_threshold}; then
    echo "=== Connection Errors found in internal traffic exceeded ${conn_error_threshold}% threshold ==="
    failed=true
  else
    echo "=== Errors found in internal traffic is within threshold ==="
  fi
  if [[ -n "${failed}" ]]; then return 1; fi
}
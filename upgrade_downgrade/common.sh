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
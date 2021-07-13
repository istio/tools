#!/usr/bin/env bash

# Copyright 2020 Istio Authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

# verifyResponseSet collects different responses from
# running a given command at most ${1} times with ${2} sleep between each run
# and verifies that the response size reaching the specified size ${3}.
# For example, verifyResponseSet 10 1 4 myFunc param1 param2
# runs "myFunc param1 param2" up to 10 times with 1 second sleep in between,
# until the response size reaching 4.
verifyResponseSet() {
    local runs=${1}
    local sleep_sec=${2}
    local size=${3}
    local n=0
    local k=""
    declare -A vals
    shift
    shift
    shift
    while (( n < runs ))
    do
        echo "RUNNING $*"
        k=$("${@}")
        vals["$k"]=1
        if [[ ${#vals[@]} -ge ${size} ]]; then
            break
        fi
        n=$(( n+1 ))
        echo "Tried $n times, sleeping ${sleep_sec} seconds and retrying..."
        sleep "${sleep_sec}"
    done
    if (( n == runs ))
    then
        exitWithErr "$* does not have a response size ${size} after running ${runs} times."
    fi
    echo "Succeeded at $*"
}

# verifyResponses verify responses from
# running a given command containing an expected result.
# ${1}: number of times running the given command.
# ${2}: sleep interval between each run.
# For example, verifyResponses 10 1 "success" myFunc param1 param2
# runs "myFunc param1 param2" 10 times with 1 second sleep in between,
# and verifies the response of each run contains "success".
verifyResponses() {
    local runs=${1}
    local sleep_sec=${2}
    local exp_str=${3}
    local n=0
    local resp=""
    shift
    shift
    shift

    while (( n < runs ))
    do
        echo "RUNNING $*"
        resp=$("${@}" 2>&1)
        arr=()
        while read -r line; do
           arr+=("$line")
        done <<< "$resp"
        contain="false"
        for line in "${arr[@]}"; do
            if [[ ${line} = *"${exp_str}"* ]]; then
                contain="true"
            fi
        done
        if [[ "${contain}" = "false" ]]; then
            break
        fi
        n=$(( n+1 ))
        echo "Ran $n times, sleeping ${sleep_sec} seconds and run again..."
        sleep "${sleep_sec}"
    done

    if (( n < runs ))
    then
        exitWithErr "$* does not have expected response when running ${runs} times."
    fi
    echo "Succeeded at $*"
}

# Parameter 1: namespace
# Parameter 2: cluster context
# Parameter 3: expected container running status (e.g., 1/1, 2/2, and etc).
waitForPodsInContextReady() {
    echo "Waiting for pods to be ready in ${1} of context ${2} ..."
    retryCmd 10 600 _waitForPodsInContextReady "${1}" "${2}" "${3}"
    echo "All pods ready."
}

# Parameter 1: namespace
# Parameter 2: cluster context
# Parameter 3: expected container running status (e.g., 1/1, 2/2, and etc).
_waitForPodsInContextReady() {
    pods_str=$(kubectl -n "${1}" --context="${2}" get pods | tail -n +2 )
    arr=()
    while read -r line; do
       arr+=("$line")
    done <<< "$pods_str"

    ready="true"
    for line in "${arr[@]}"; do
        if [[ ${line} != *"${3}"*"Running"* && ${line} != *"Completed"* ]]; then
            ready="false"
        fi
    done
    if [  "${ready}" = "true" ]; then
        return 0
    fi

    echo "${pods_str}"
    return 1
}

# retryCmd runs a given command with retries.
# Parameter 1: retry interval.
# Parameter 2: max retry time.
# Parameter 3: command to run
# For example, "retryCmd 1 10 cmd cmdParam1" will run "cmd cmdParam1" with 1 second interval until
# either the cmd succeeds or 10 seconds have reached.
retryCmd() {
    local retry_interval=${1}
    local max_retry_time=${2}
    shift
    shift

    local first_time=${SECONDS}
    while (( SECONDS - first_time <  max_retry_time )); do
        echo "Run $*"
        "${@}" && break
        echo "Retry after ${retry_interval} seconds ..."
        sleep "${retry_interval}"
    done

    if (( SECONDS - first_time >=  max_retry_time )); then
        exitWithErr "$* failed with retry ${max_retry_time} seconds."
    fi
    echo "Succeeded at $*"
}

# Output error message and exit.
exitWithErr() {
    echo "$*" 1>&2 ; exit 1;
}
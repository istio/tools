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

echo_and_run() { echo "# RUNNING $*" ; "$@" ; }
echo_and_run_quiet() { echo "# RUNNING(quiet) $*" ; "$@" > /dev/null 2>&1 ; }
echo_and_run_or_die() { echo "# RUNNING $*" ; "$@" || die "failed!" ; }

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

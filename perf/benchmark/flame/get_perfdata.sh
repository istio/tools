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

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

WD=$(dirname "${0}")
WD=$(cd "${WD}" && pwd)

PERF_FILENAME=${1:?"perf filename is missing"}
PERF_DURATION=${2:?"perf duration is missing"}
SAMPLE_FREQUENCY=${3:-"99"}

PID=$(pgrep envoy)

# This is specific to the kernel version
# example: /usr/lib/linux-tools-4.4.0-131/perf
# provided by `linux-tools-generic`
PERFDIR=$(find /usr/lib -name 'linux-tools-*' -type d | head -n 1)
if [[ -z "${PERFDIR}" ]]; then
    echo "Missing perf tool. Install apt-get install linux-tools-generic"
    exit 1
fi

PERF="${PERFDIR}/perf"

"${PERF}" record -o "${WD}/${PERF_FILENAME}" -F "${SAMPLE_FREQUENCY}" -p "${PID}" -g -- sleep "${PERF_DURATION}"
"${PERF}" script -i "${WD}/${PERF_FILENAME}" --demangle > "${WD}/${PERF_FILENAME}.perf"

echo "Wrote ${WD}/${PERF_FILENAME}.perf"

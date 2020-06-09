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

FLAMEDIR="${WD}/FlameGraph"

if ! command -v c++filt > /dev/null; then
    echo "Install c++filt to demangle symbols"
    exit 1
fi

cd "${WD}" || exit 1

if [[ ! -d ${FLAMEDIR} ]]; then
    echo "Cloning FlameGraph repo in ${WD}"
    git clone https://github.com/brendangregg/FlameGraph
fi

# Given output of `perf script` produce a flamegraph
FILE=${1:?"get_perfdata script output"}
FILENAME=$(basename "${FILE}")
BASE=$(echo "${FILENAME}" | cut -d '.' -f 1)
SVGNAME="${BASE}.svg"

mkdir -p "${WD}/flameoutput"
"${FLAMEDIR}/stackcollapse-perf.pl" "${FILE}" | c++filt -n | "${FLAMEDIR}/flamegraph.pl" --cp > "./flameoutput/${SVGNAME}"

echo "Wrote CPU flame graph for istio-proxy ${SVGNAME}"


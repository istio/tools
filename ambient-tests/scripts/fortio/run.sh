#! /bin/bash

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

# Run netperf performance benchmarks

set -eux

# shellcheck disable=SC1091
source scripts/config.sh

mkdir -p "$FORTIO_RESULTS"

function run-tests() {
    client_ns=$1
    server_ns=$2

    for _ in $(seq "$N_RUNS")
    do
        # shellcheck disable=SC2086
        kubectl exec -it -n "$client_ns" deploy/client \
            -- fortio load $FORTIO_SERIAL_HTTP_ARGS -json serial.json "http://$BENCHMARK_SERVER.$server_ns:8080" \
            > /dev/null
        echo "$client_ns:$server_ns" 
        kubectl exec -it -n "$client_ns" deploy/client -- cat serial.json
        echo "$TEST_RUN_SEPARATOR"
    done >> "$FORTIO_RESULTS/serial"

    for _ in $(seq "$N_RUNS")
    do
        # shellcheck disable=SC2086
        kubectl exec -it -n "$client_ns" deploy/client \
            -- fortio load $FORTIO_PARALLEL_HTTP_ARGS -json parallel.json "http://$BENCHMARK_SERVER.$server_ns:8080" \
            > /dev/null
        echo "$client_ns:$server_ns"
        kubectl exec -it -n "$client_ns" deploy/client -- cat parallel.json
        echo "$TEST_RUN_SEPARATOR"
    done >> "$FORTIO_RESULTS/parallel"
}

true > "$FORTIO_RESULTS/serial"
true > "$FORTIO_RESULTS/parallel"
run-tests "$NS_AMBIENT"  "$NS_AMBIENT"
run-tests "$NS_SIDECAR"  "$NS_SIDECAR"
run-tests "$NS_NO_MESH"  "$NS_NO_MESH"
run-tests "$NS_WAYPOINT" "$NS_WAYPOINT"

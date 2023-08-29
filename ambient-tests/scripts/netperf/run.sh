#!/usr/bin/env bash

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

set -eux

# shellcheck disable=SC1091
source scripts/config.sh

INTER_TEST_SLEEP=0.1s

mkdir -p "$NETPERF_RESULTS"

# First argument is the client namespace
# Second argument is the server namespace
function run-tests() {
    # give values names
    client_ns=$1
    server_ns=$2
    for _ in $(seq "$N_RUNS")
    do
        # shellcheck disable=SC2086
        kubectl exec "deploy/$BENCHMARK_CLIENT" -n "$client_ns" \
        -- netperf $NETPERF_GLOBAL_ARGS -H "$BENCHMARK_SERVER.$server_ns" -t TCP_STREAM \
        -- $NETPERF_TEST_ARGS
        echo "NAMESPACES=$client_ns:$server_ns"
        echo "$TEST_RUN_SEPARATOR"
        sleep "$INTER_TEST_SLEEP"
    done >> "$NETPERF_RESULTS/TCP_STREAM"

    for _ in $(seq "$N_RUNS")
    do
        # shellcheck disable=SC2086
        kubectl exec "deploy/$BENCHMARK_CLIENT" -n "$client_ns" \
        -- netperf $NETPERF_GLOBAL_ARGS -H "$BENCHMARK_SERVER.$server_ns" -t TCP_CRR \
        -- $NETPERF_TEST_ARGS $NETPERF_RR_ARGS 
        echo "NAMESPACES=$client_ns:$server_ns"
        echo "$TEST_RUN_SEPARATOR"
        sleep "$INTER_TEST_SLEEP"
    done >> "$NETPERF_RESULTS/TCP_CRR"

    for _ in $(seq "$N_RUNS")
    do
        # shellcheck disable=SC2086
        kubectl exec "deploy/$BENCHMARK_CLIENT" -n "$client_ns" \
        -- netperf $NETPERF_GLOBAL_ARGS -H "$BENCHMARK_SERVER.$server_ns" -t TCP_RR \
        -- $NETPERF_TEST_ARGS $NETPERF_RR_ARGS 
        echo "NAMESPACES=$client_ns:$server_ns"
        echo "$TEST_RUN_SEPARATOR"
        sleep "$INTER_TEST_SLEEP"
    done >> "$NETPERF_RESULTS/TCP_RR" 
}

# clear output files
true > "$NETPERF_RESULTS/TCP_STREAM"
true > "$NETPERF_RESULTS/TCP_CRR"
true > "$NETPERF_RESULTS/TCP_RR"

run-tests "$NS_AMBIENT"  "$NS_AMBIENT" 
run-tests "$NS_NO_MESH"  "$NS_NO_MESH" 
run-tests "$NS_SIDECAR"  "$NS_SIDECAR"   
run-tests "$NS_WAYPOINT" "$NS_WAYPOINT"
# For cross-mesh tests
# run-tests "$NS_SIDECAR" "$NS_AMBIENT"   


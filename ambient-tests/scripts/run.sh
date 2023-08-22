#! /bin/bash
# Run performance benchmarks

set -eux

# shellcheck disable=SC1091
source scripts/config

INTER_TEST_SLEEP=0.1s

mkdir -p "$RESULTS"

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
        -- netperf $GLOBAL_ARGS -H "$BENCHMARK_SERVER.$server_ns" -t TCP_STREAM \
        -- $TEST_ARGS
        echo "NAMESPACES=$client_ns:$server_ns"
        echo "$TEST_RUN_SEPARATOR"
        sleep "$INTER_TEST_SLEEP"
    done >> "$RESULTS/TCP_STREAM"

    for _ in $(seq "$N_RUNS")
    do
        # shellcheck disable=SC2086
        kubectl exec "deploy/$BENCHMARK_CLIENT" -n "$client_ns" \
        -- netperf $GLOBAL_ARGS -H "$BENCHMARK_SERVER.$server_ns" -t TCP_CRR \
        -- $TEST_ARGS $RR_ARGS 
        echo "NAMESPACES=$client_ns:$server_ns"
        echo "$TEST_RUN_SEPARATOR"
        sleep "$INTER_TEST_SLEEP"
    done >> "$RESULTS/TCP_CRR"

    for _ in $(seq "$N_RUNS")
    do
        # shellcheck disable=SC2086
        kubectl exec "deploy/$BENCHMARK_CLIENT" -n "$client_ns" \
        -- netperf $GLOBAL_ARGS -H "$BENCHMARK_SERVER.$server_ns" -t TCP_RR \
        -- $TEST_ARGS $RR_ARGS 
        echo "NAMESPACES=$client_ns:$server_ns"
        echo "$TEST_RUN_SEPARATOR"
        sleep "$INTER_TEST_SLEEP"
    done >> "$RESULTS/TCP_RR" 
}

# clear output files
true > "$RESULTS/TCP_STREAM"
true > "$RESULTS/TCP_CRR"
true > "$RESULTS/TCP_RR"

run-tests "$NS_AMBIENT" "$NS_AMBIENT" 
run-tests "$NS_NO_MESH" "$NS_NO_MESH" 
run-tests "$NS_SIDECAR" "$NS_SIDECAR"   
# For cross-mesh tests
# run-tests "$NS_SIDECAR" "$NS_AMBIENT"   


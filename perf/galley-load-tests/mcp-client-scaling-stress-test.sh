#!/bin/bash

set -o errexit

MAX_REPLICAS=${MAX_REPLICAS:-100}
MIN_REPLICAS=${MIN_REPLICAS:-10}

# setup port forwarding to galley to query memory and cpu profiles
GALLEY_POD=$(kubectl -n istio-system get pod -listio=galley --field-selector=status.phase=Running -o jsonpath='{.items[0].metadata.name}')
echo "Profiling "${GALLEY_POD}

kubectl -n istio-system port-forward ${GALLEY_POD} 9094 >/dev/null &
GALLEY_PID=$!
on_exit() {
    kill ${GALLEY_PID}
}
trap on_exit EXIT

sleep 1 # give port forwarding a chance to start

num_goroutines() {
    echo $(curl -s http://localhost:9094/debug/pprof/goroutine?debug=1 | grep ^"goroutine profile"|cut -f4 -d' ')
}

while true; do
    # scale up
    for replicas in $(seq ${MIN_REPLICAS} 1 ${MAX_REPLICAS}); do
        kubectl -n istio-system scale deployment istio-pilot --replicas ${replicas}
        kubectl -n istio-system rollout status deployment istio-pilot

        echo "[up]" $(num_goroutines) goroutines for ${REPLICAS} replicas
    done

    # scale down
    for replicas in $(seq ${MAX_REPLICAS} -1 ${MIN_REPLICAS}); do
        kubectl -n istio-system scale deployment istio-pilot --replicas ${replicas}
        kubectl -n istio-system rollout status deployment istio-pilot

        echo "[down]" $(num_goroutines) goroutines for ${REPLICAS} replicas
    done
done

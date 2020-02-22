#!/bin/bash

function cleanup() {
    kill -s TERM $PID
    wait $PID
}

set -e 
set -x
python3 prom.py --http-port 64634 &
PID=$!
trap cleanup EXIT
sleep 2
curl --silent 127.0.0.1:64634 | grep cpu_stats_interrupts

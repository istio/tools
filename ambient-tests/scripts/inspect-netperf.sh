#! /bin/bas
# Script to run CRR tests with a while recording data and save the pcap file locally.
# This will terminal all tcpdump instances in the pod's pid namespace.

set -eux

source scripts/config.sh

PCAP_FN=trace.pcap

client_ns="$1"
server_ns="$2"

kubectl exec "deploy/$BENCHMARK_CLIENT" -n "$client_ns" -- tcpdump -w "$PCAP_FN" -i lo &

TCPDUMP_PID=$!
sleep 1
echo $TCPDUMP_PID

kubectl exec "deploy/$BENCHMARK_CLIENT" -n "$client_ns" \
-- netperf -H $GLOBAL_ARGS -H "$BENCHMARK_SERVER.$server_ns" -t TCP_CRR \
-- $TEST_ARGS -r 100 > /dev/null

kubectl exec "deploy/$BENCHMARK_CLIENT" -n "$client_ns" -- pkill tcpdump

pod_name=$(kubectl get pod -l app=client -n "$client_ns" -o json | jq '.items[0].metadata.name' | tr -d '"')
kubectl cp "$client_ns/$pod_name:$PCAP_FN" ./"$PCAP_FN"


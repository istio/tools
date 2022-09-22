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

BENCHMARK_DIR=$(dirname "$0")/../../..
CONFIG_DIR=$(dirname "$0")
FORTIOSERVER=$(kubectl get pods -n twopods-istio --selector=app=fortioserver --output=jsonpath={.items..metadata.name})
PROVIDER=$(kubectl get services -n twopods-istio --selector=app=ext-authz --output=jsonpath={.items..spec.clusterIP})

# In case the policy has benn applied, try to delete first
kubectl delete -n twopods-istio -f ${CONFIG_DIR}/policy.yaml

# client to server, without ext-authz
python3 ${BENCHMARK_DIR}/runner/runner.py \
    --config_file=${CONFIG_DIR}/no_ext-authz_multi_conn_latency.yaml
python3 ${BENCHMARK_DIR}/runner/runner.py \
    --config_file=${CONFIG_DIR}/no_ext-authz_multi_qps_latency.yaml

kubectl apply -n twopods-istio -f ${CONFIG_DIR}/policy.yaml

# client to server, with ext-authz
python3 ${BENCHMARK_DIR}/runner/runner.py \
    --config_file=${CONFIG_DIR}/with_ext-authz_multi_conn_latency.yaml
python3 ${BENCHMARK_DIR}/runner/runner.py \
    --config_file=${CONFIG_DIR}/with_ext-authz_multi_qps_latency.yaml

# server to ext-authz provider
for conn in 32 64 128 256 512
do
    kubectl -n twopods-istio exec ${FORTIOSERVER}  \
        -- fortio load -H=x-ext-authz:allow  -jitter=True -c $conn -qps 500 \
        -t 240s -a -r 0.001 -httpbufferkb=128 \
        -labels qps_500_c_${conn}_v2-stats-nullvm_to-ext-authz_both http://${PROVIDER}:8000/echo?size=1024
done

for qps in 500 600 700 800 900 1000
do
    kubectl -n twopods-istio exec ${FORTIOSERVER}  \
        -- fortio load -H=x-ext-authz:allow  -jitter=True -c 64 -qps $qps \
        -t 240s -a -r 0.001 -httpbufferkb=128 \
        -labels qps_${qps}_c_64_1024_v2-stats-nullvm_to-ext-authz_both http://${PROVIDER}:8000/echo?size=1024
done
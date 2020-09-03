#!/bin/bash

# Copyright Istio Authors

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#    http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "Creating tests for authorizationPolicy with a variable number of policies each with 1 sourceIP rule"
echo "Running each test in small load (conn=8, qps=100), medium load (conn=32, qps=500), and large load (conn=64, qps=1000)"

echo '
{
    "authZ":
    {
        "numPolicies":10,
        "numSourceIP":1
    }
}' > authZ10SourceIP.json
go run ../generate_policies.go ../generate.go ../jwt.go -configFile="authZ10SourceIP.json" > authZ10SourceIP.yaml
echo "Generated 10 authZ policy with 1 sourceIP rule each"
kubectl apply -f authZ10SourceIP.yaml
echo "Running variable number of policies"
echo "Running perf test with conn=8 and qps=100"
pipenv run python3 ../../../runner/runner.py --conn 8 --qps 100 --baseline --duration 240 --load_gen_type=nighthawk --telemetry_mode=none
echo "Running perf test with conn=32 and qps=500"
pipenv run python3 ../../../runner/runner.py --conn 32 --qps 500 --baseline --duration 240 --load_gen_type=nighthawk --telemetry_mode=none
echo "Running perf test with conn=64 and qps=1000"
pipenv run python3 ../../../runner/runner.py --conn 64 --qps 1000 --baseline --duration 240 --load_gen_type=nighthawk --telemetry_mode=none
echo "Removing policies"
kubectl delete -f authZ10SourceIP.yaml
rm authZ10SourceIP.json
rm authZ10SourceIP.yaml

echo '
{
    "authZ":
    {
        "numPolicies":1000,
        "numSourceIP":1
    }
}' > authZ1000SourceIP.json
go run ../generate_policies.go ../generate.go ../jwt.go -configFile="authZ1000SourceIP.json" > authZ1000SourceIP.yaml
echo "Generated 1000 authZ policy with 1 sourceIP rule each"
kubectl apply -f authZ1000SourceIP.yaml
echo "Running perf test with conn=8 and qps=100"
pipenv run python3 ../../../runner/runner.py --conn 8 --qps 100 --baseline --duration 240 --load_gen_type=nighthawk --telemetry_mode=none
echo "Running perf test with conn=32 and qps=500"
pipenv run python3 ../../../runner/runner.py --conn 32 --qps 500 --baseline --duration 240 --load_gen_type=nighthawk --telemetry_mode=none
echo "Running perf test with conn=64 and qps=1000"
pipenv run python3 ../../../runner/runner.py --conn 64 --qps 1000 --baseline --duration 240 --load_gen_type=nighthawk --telemetry_mode=none
echo "Removing policies"
kubectl delete -f authZ1000SourceIP.yaml
rm authZ1000SourceIP.json
rm authZ1000SourceIP.yaml

echo "Fetching data"
FORTIO_CLIENT_URL=http://$(kubectl get services -n twopods-istio fortioclient -o jsonpath="{.status.loadBalancer.ingress[0].ip}"):9076
kubectl -n istio-prometheus port-forward svc/istio-prometheus 9090:9090 &
PROMETHEUS_URL=http://localhost:9090
python3 ./../../../runner/fortio.py "$FORTIO_CLIENT_URL" --prometheus=$PROMETHEUS_URL --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_istio_proxy_fortioclient,cpu_mili_avg_istio_proxy_fortioserver,cpu_mili_avg_istio_proxy_istio-ingressgateway,mem_Mi_avg_istio_proxy_fortioclient,mem_Mi_avg_istio_proxy_fortioserver,mem_Mi_avg_istio_proxy_istio-ingressgateway

echo "Cleanup started"
kubectl delete --all pods --namespace=twopods-istio
echo "Cleanup finished, data collected"

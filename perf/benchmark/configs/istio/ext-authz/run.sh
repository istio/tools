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

CONFIG_DIR=$(dirname "$0")
FORTIOCLIENT=$(kubectl get pods -n twopods-istio --selector=app=fortioclient --output=jsonpath="{.items[0].metadata.name}")
PROVIDER=$(kubectl get services -n twopods-istio --selector=app=ext-authz --output=jsonpath="{.items[0].spec.clusterIP}")

# In case the policy has benn applied, try to delete first
kubectl delete -n twopods-istio -f "${CONFIG_DIR}/policy.yaml" || true

# client to server, without ext-authz
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -jitter=True -c 8 -qps 100 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_100_c_8_without-ext-authz_small http://fortioserver:8080/echo?size=1024
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -jitter=True -c 32 -qps 500 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_500_c_32_without-ext-authz_medium http://fortioserver:8080/echo?size=1024
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -jitter=True -c 64 -qps 1000 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_1000_c_64_without-ext-authz_large http://fortioserver:8080/echo?size=1024

kubectl apply -n twopods-istio -f "${CONFIG_DIR}/policy.yaml"

# client to server, with ext-authz
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -H=x-ext-authz:allow -jitter=True -c 8 -qps 100 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_100_c_8_with-ext-authz_small http://fortioserver:8080/echo?size=1024
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -H=x-ext-authz:allow -jitter=True -c 32 -qps 500 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_500_c_32_with-ext-authz_medium http://fortioserver:8080/echo?size=1024
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -H=x-ext-authz:allow -jitter=True -c 64 -qps 1000 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_1000_c_64_with-ext-authz_large http://fortioserver:8080/echo?size=1024

# client to ext-authz provider
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -H=x-ext-authz:allow -jitter=True -c 8 -qps 100 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_100_c_8_to-ext-authz_small "http://${PROVIDER}:8000/echo?size=1024"
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -H=x-ext-authz:allow -jitter=True -c 32 -qps 500 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_500_c_32_to-ext-authz_medium "http://${PROVIDER}:8000/echo?size=1024"
kubectl -n twopods-istio exec "${FORTIOCLIENT}"  \
    -- fortio load -H=x-ext-authz:allow -jitter=True -c 64 -qps 1000 \
    -t 300s -a -r 0.001 -httpbufferkb=128 \
    -labels qps_1000_c_64_to-ext-authz_large "http://${PROVIDER}:8000/echo?size=1024"


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

set -ex
NUM=${NUM:?"specify the number of httpbin and sleep workloads"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

if ! istioctl version; then 
    exit
fi

timestamp=$(date +"%Y-%m-%d-%H-%M-%S")
staticns=static-${timestamp}
dynamicns=dynamic-${timestamp}
dynamicworkloadlife=180

kubectl create ns "${staticns}" --cluster "${CLUSTER}"
kubectl create ns "${dynamicns}" --cluster "${CLUSTER}"
kubectl label namespace "${staticns}" istio-injection=enabled --cluster "${CLUSTER}"
kubectl label namespace "${dynamicns}" istio-injection=enabled --cluster "${CLUSTER}"
kubectl get ns --show-labels --cluster "${CLUSTER}"
kubectl apply -n "${staticns}" -f workload.yaml --cluster "${CLUSTER}"
kubectl apply -n "${dynamicns}" -f workload.yaml --cluster "${CLUSTER}"
kubectl -n "${staticns}" scale deployment httpbin --replicas="${NUM}" --cluster "${CLUSTER}"
kubectl -n "${staticns}" scale deployment sleep --replicas="${NUM}" --cluster "${CLUSTER}"
kubectl -n "${dynamicns}" scale deployment httpbin --replicas="${NUM}" --cluster "${CLUSTER}"
kubectl -n "${dynamicns}" scale deployment sleep --replicas="${NUM}" --cluster "${CLUSTER}"

helm -n "${dynamicns}" template \
    --set Namespace="${dynamicns}" \
    --set Num="${NUM}" \
    --set WorkloadLife="${dynamicworkloadlife}" \
          . > auto-rotate.yaml
kubectl apply -n "${dynamicns}" -f auto-rotate.yaml --cluster "${CLUSTER}"

# echo "Wait 60 seconds for the deployment to be ready ..."
# sleep 60

rm auto-rotate.yaml

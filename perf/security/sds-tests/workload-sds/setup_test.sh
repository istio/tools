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

# specify the Istio release version, e.g., release-1.1-20190208-09-16
release_version=$1
# specify the Istio release type, daily, release, pre-release
release_type=$2

# shellcheck disable=SC1091	
source ../../utils/get_release.sh

function download_istio() {
    echo "Download istioctl from release version ${release_version}, release type ${release_type}"
    # shellcheck disable=SC2086
    wd=$(dirname $0)/tmp
    if [[ ! -d "${wd}" ]]; then	
        # shellcheck disable=SC2086	
        mkdir $wd
    fi
    get_release_url "$release_type" "$release_version"
    # shellcheck disable=SC2154	
    if [[ -z "$release_url" ]]; then	
        return 1	
    fi
    curl -JLo "$wd/istio-${release_version}.tar.gz" "${release_url}"	
    # shellcheck disable=SC2086	
    tar xfz ${wd}/istio-${release_version}.tar.gz -C $wd
    export PATH=$PWD/tmp/istio-${release_version}/bin:$PATH
}

if [[ -n $release_version ]] && [[ -n $release_type ]]; then 
    download_istio
    return_code=$?
    if [ "$return_code" -eq 1 ]; then
        echo "failed in downloading istio, exit"
        exit
    fi
fi

if ! istioctl version; then 
    echo "istioctl is not installed or invalid istioctl version"
    exit
fi

timestamp=$(date +"%Y-%m-%d-%H-%M-%S")
staticns=static-${timestamp}
dynamicns=dynamic-${timestamp}
# redeploy workloads in dynamic namespace every 180 seconds
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

# echo "Wait 10 seconds for the deployment to be ready ..."
sleep 10

rm auto-rotate.yaml

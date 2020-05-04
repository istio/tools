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
NUM=${NUM:?"specify the number of gateway"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

# specify the Istio release version, e.g., release-1.1-20190208-09-16
release_version=$1
# specify the Istio release type, daily, release, pre-release
release_type=$2

# shellcheck disable=SC1091	
source ../../utils/get_release.sh
# shellcheck disable=SC1091
source setup_gateway.sh
# shellcheck disable=SC1091
source setup_virtualservice.sh
# shellcheck disable=SC1091
source setup_client.sh

wd=""
# download istio release package into a tmp folder
function download_istio() {
    echo "Download istioctl from release version ${release_version}, release type ${release_type}"
    # shellcheck disable=SC2086
   
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

# prepare istioctl exit if istioctl does not exist
function prepare_istioctl() {
    # shellcheck disable=SC2086
    wd=$(dirname $0)/tmp
    if [[ ! -d "${wd}" ]]; then	
        # shellcheck disable=SC2086	
        mkdir $wd
    fi
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
}

secure_ingress_port=""
ingress_host=""
function set_ingress_host_port() {
    secure_ingress_port=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].port}')
    ingress_host=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    echo "secure ingress port: ${secure_ingress_port}"
    echo "ingress host: ${ingress_host}"
}

function prepare_ingress_secret() {
    # shellcheck disable=SC2086
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout ${wd}/example.com.key -out ${wd}/example.com.crt
    # shellcheck disable=SC2086
    kubectl create -n istio-system secret generic ingress-root --from-file=key=${wd}/example.com.key --from-file=cert=${wd}/example.com.crt
    # shellcheck disable=SC2004
    for ((id=1; id<=${NUM}; id++)); do
        credential_name="httpbin-credential-${id}"
        # shellcheck disable=SC2086
        openssl req -out ${wd}/httpbin.example.com.csr -newkey rsa:2048 -nodes -keyout ${wd}/httpbin.example.com.key -subj "/CN=httpbin.example.com/O=httpbin organization"
        # shellcheck disable=SC2086
        openssl x509 -req -days 365 -CA ${wd}/example.com.crt -CAkey ${wd}/example.com.key -set_serial 0 -in ${wd}/httpbin.example.com.csr -out ${wd}/httpbin.example.com.crt
        # shellcheck disable=SC2086
        kubectl create -n istio-system secret generic "${credential_name}" --from-file=key=${wd}/httpbin.example.com.key --from-file=cert=${wd}/httpbin.example.com.crt
    done
}

timestamp=$(date +"%Y-%m-%d-%H-%M-%S")
testns=httpbin-${timestamp}
function deploy_httpbin() {
    kubectl create ns "${testns}" --cluster "${CLUSTER}"
    kubectl apply -n "${testns}" --cluster "${CLUSTER}" -f https://raw.githubusercontent.com/istio/istio/release-1.5/samples/httpbin/httpbin.yaml
}

function deploy_gateways() {
    # shellcheck disable=SC2004
    for ((id=1; id<=${NUM}; id++)); do
        deploy_gateway $id "${testns}" "${CLUSTER}"
        deploy_virtualservice $id "${testns}" "${CLUSTER}"
    done
}

function deploy_clients() {
    # shellcheck disable=SC2004
    for ((id=1; id<=${NUM}; id++)); do
        deploy_sleep $id "${testns}" "${CLUSTER}" "${wd}"
    done
}

function check_access() {
    # shellcheck disable=SC2004
    for ((id=1; id<=${NUM}; id++)); do
        host="httpbin-${id}.example.com"
        # shellcheck disable=SC2153
        url="https://httpbin-${id}.example.com:${SECURE_INGRESS_PORT}/status/418"
        # shellcheck disable=SC2153
        curl -v -HHost:"${host}" --resolve "${host}":"${SECURE_INGRESS_PORT}":"${INGRESS_HOST}" --cacert "${wd}"/example.com.crt "${url}"
    done
}

prepare_istioctl
set_ingress_host_port
prepare_ingress_secret
deploy_httpbin
deploy_gateways
check_access

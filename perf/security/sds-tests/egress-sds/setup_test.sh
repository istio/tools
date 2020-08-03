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
NUM=${NUM:?"specify the number of gateways"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}
kc="kubectl --cluster ${CLUSTER}"
# specify the Istio release version, e.g., release-1.1-20190208-09-16
release_version=$1
# specify the Istio release type, daily, release, pre-release
release_type=$2

# shellcheck disable=SC1091
source ../../utils/get_release.sh
# shellcheck disable=SC1091
source setup_nginx.sh
# shellcheck disable=SC1091
source setup_client.sh
# shellcheck disable=SC1091
source setup_gateway.sh
# shellcheck disable=SC1091
source setup_virtualservice.sh
# shellcheck disable=SC1091
source setup_destinationrule.sh

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

function prepare_root_ca() {
    # shellcheck disable=SC2086
    openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=example Inc./CN=example.com' -keyout ${wd}/example.com.key -out ${wd}/example.com.crt
}

testns=clientns

function prepare_namespace() {
    # shellcheck disable=SC2086
    ${kc} create namespace mesh-external
    ${kc} create -n mesh-external secret generic nginx-ca-certs --from-file="${wd}"/example.com.crt
    ${kc} create namespace ${testns}
    ${kc} label namespace ${testns} istio-injection=enabled
}

function deploy_nginx() {
    for ((id=1; id<="${NUM}"; id++)); do
        credential_name="nginx-server-certs-${id}"
        # shellcheck disable=SC2086
        openssl req -out ${wd}/my-nginx.mesh-external.svc.cluster.local.csr -newkey rsa:2048 -nodes -keyout ${wd}/mesh-external.svc.cluster.local.key -subj "/CN=my-nginx-${id}.mesh-external.svc.cluster.local/O=some organization"
        # shellcheck disable=SC2086
        openssl x509 -req -days 365 -CA ${wd}/example.com.crt -CAkey ${wd}/example.com.key -set_serial 0 -in ${wd}/my-nginx.mesh-external.svc.cluster.local.csr -out ${wd}/my-nginx.mesh-external.svc.cluster.local.crt
        # shellcheck disable=SC2086
        ${kc} create -n mesh-external secret tls "${credential_name}" --key ${wd}/mesh-external.svc.cluster.local.key --cert=${wd}/my-nginx.mesh-external.svc.cluster.local.crt

        setup_nginx $id "mesh-external" "${CLUSTER}" "${wd}"
    done
}

function deploy_clients() {
    # shellcheck disable=SC2004
    for ((id=1; id<=${NUM}; id++)); do
        deploy_sleep $id "${testns}" "${CLUSTER}" "${wd}"
    done
}

function deploy_gateways() {
    # shellcheck disable=SC2004
    for ((id=1; id<=${NUM}; id++)); do
        setup_gateways $id "${testns}" "${CLUSTER}"
    done
}

function deploy_virtualservice() {
    # shellcheck disable=SC2004
    for ((id=1; id<=${NUM}; id++)); do
        setup_virtualservices $id "${testns}" "${CLUSTER}"
    done
}

function deploy_destinationrule() {
    # shellcheck disable=SC2004
    for ((id=1; id<=${NUM}; id++)); do
        credential_name="client-credential-${id}"
        # shellcheck disable=SC2086
        openssl req -out ${wd}/client.example.com.csr -newkey rsa:2048 -nodes -keyout ${wd}/client.example.com.key -subj "/CN=client.example.com/O=client organization"
        # shellcheck disable=SC2086
        openssl x509 -req -days 365 -CA ${wd}/example.com.crt -CAkey ${wd}/example.com.key -set_serial 1 -in ${wd}/client.example.com.csr -out ${wd}/client.example.com.crt
        # shellcheck disable=SC2086
        ${kc} create -n istio-system secret generic "${credential_name}" --from-file=tls.key=${wd}/client.example.com.key --from-file=tls.crt=${wd}/client.example.com.crt --from-file=ca.crt=${wd}/example.com.crt

        setup_destinationrules $id "${CLUSTER}"
    done
}

prepare_istioctl
prepare_root_ca
prepare_namespace
deploy_nginx
deploy_gateways
deploy_virtualservice
deploy_destinationrule
deploy_clients

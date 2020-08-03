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

function cleanup() {
  ${kc} delete ns mesh-external
  ${kc} delete ns clientns
  for s in $(kubectl -n istio-system get secrets -oname | grep "client-credential-*")
  do
    kubectl -n istio-system delete "${s}"
  done
  for s in $(kubectl -n istio-system get secrets -oname | grep "nginx-server-certs-*")
  do
    kubectl -n istio-system delete "${s}"
  done
}

cleanup

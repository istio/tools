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

WD=$(dirname $0)
WD=$(cd $WD; pwd)

NAMESPACE="${NAMESPACE:-"istio-stability-gateway-bouncer"}"
${WD}/../setup_test.sh "gateway-bouncer" "--set namespace=${NAMESPACE}"

if [ -z ${DRY_RUN}; then
  # Waiting until LoadBalancer is created and retrieving the assigned
  # external IP address.
  while : ; do
    INGRESS_IP=$(kubectl -n ${NAMESPACE} \
      get service istio-ingress-${NAMESPACE} \
      -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

    if [[ -z "${INGRESS_IP}" ]]; then
      sleep 5s
    else
      break
    fi
  done

  # Populating a ConfigMap with the external IP address and restarting the
  # client to pick up the new version of the ConfigMap.
  kubectl -n ${NAMESPACE} delete configmap fortio-client-config
  kubectl -n ${NAMESPACE} create configmap fortio-client-config \
    --from-literal=external_addr=${INGRESS_IP}
  kubectl -n ${NAMESPACE} patch deployment fortio-client \
    -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"date\":\"`date +'%s'`\"}}}}}"
fi
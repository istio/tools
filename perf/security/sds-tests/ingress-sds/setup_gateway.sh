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

function deploy_gateway() {
    local id="${1:?"please specify the gateway id"}"
    local ns="${2:?"please specify the namespace"}"
    local cs="${3:?"please specify the cluster"}"
    local gateway_name="mygateway-${id}"
    local credential_name="httpbin-credential-${id}"
    local host="httpbin-${id}.example.com"

    cat <<-EOF | kubectl apply -n "${ns}" --cluster "${cs}" -f -
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: "${gateway_name}"
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    tls:
      mode: SIMPLE
      credentialName: "${credential_name}"
    hosts:
    - "${host}"
EOF
}

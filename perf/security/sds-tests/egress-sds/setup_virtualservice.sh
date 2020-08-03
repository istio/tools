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

function setup_virtualservices() {
    local id="${1:?"please specify the gateway id"}"
    local ns="${2:?"please specify the namespace"}"
    local cs="${3:?"please specify the cluster"}"
    local gateway_name="istio-egressgateway-${id}"
    local vs_name="direct-nginx-through-egress-gateway-${id}"
    local subset="nginx-${id}"
    local host_gateway="istio-egressgateway.istio-system.svc.cluster.local"
    local host_nginx="my-nginx-${id}.mesh-external.svc.cluster.local"

    cat <<EOF | kubectl apply -n "${ns}" --cluster "${cs}" -f -
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: "${vs_name}"
spec:
  hosts:
  - "${host_nginx}"
  gateways:
  - "${gateway_name}"
  - mesh
  http:
  - match:
    - gateways:
      - mesh
      port: 80
    route:
    - destination:
        host: "${host_gateway}"
        subset: "${subset}"
        port:
          number: 443
      weight: 100
  - match:
    - gateways:
      - "${gateway_name}"
      port: 443
    route:
    - destination:
        host: "${host_nginx}"
        port:
          number: 443
      weight: 100
EOF
}

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

function setup_gateways() {
    local id="${1:?"please specify the gateway id"}"
    local ns="${2:?"please specify the namespace"}"
    local cs="${3:?"please specify the cluster"}"
    local gateway_name="istio-egressgateway-${id}"
    local gateway_dr="egressgateway-for-nginx-${id}"
    local subset="nginx-${id}"
    local host_gateway="istio-egressgateway.istio-system.svc.cluster.local"
    local host_nginx="my-nginx-${id}.mesh-external.svc.cluster.local"

    cat <<-EOF | kubectl apply -n "${ns}" --cluster "${cs}" -f -
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: "${gateway_name}"
spec:
  selector:
    istio: egressgateway
  servers:
  - port:
      number: 443
      name: https
      protocol: HTTPS
    hosts:
    - "${host_nginx}"
    tls:
      mode: ISTIO_MUTUAL
---
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: "${gateway_dr}"
spec:
  host: "${host_gateway}"
  subsets:
  - name: "${subset}"
    trafficPolicy:
      loadBalancer:
        simple: ROUND_ROBIN
      portLevelSettings:
      - port:
          number: 443
        tls:
          mode: ISTIO_MUTUAL
          sni: "${host_nginx}"
EOF
     sleep 2
}

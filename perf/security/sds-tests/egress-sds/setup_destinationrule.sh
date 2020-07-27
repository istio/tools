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

function setup_destinationrules() {
    local id="${1:?"please specify the gateway id"}"
    local cs="${2:?"please specify the cluster"}"
    local dr_name="originate-mtls-for-nginx-${id}"
    local host="my-nginx-${id}.mesh-external.svc.cluster.local"
    local credential_name="client-credential-${id}"

    cat <<-EOF | kubectl apply -n istio-system --cluster "${cs}" -f -
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: "${dr_name}"
spec:
  host: "${host}"
  trafficPolicy:
    loadBalancer:
      simple: ROUND_ROBIN
    portLevelSettings:
    - port:
        number: 443
      tls:
        mode: MUTUAL
        credentialName: "${credential_name}" # this must match the secret created earlier to hold client certs
        sni: "${host}"
EOF

#    istioctl experimental wait -n istio-system destinationrule "${dr_name}" --timeout=60s
}

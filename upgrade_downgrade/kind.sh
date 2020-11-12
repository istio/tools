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

function cleanup_kind_cluster() {
  NAME=istio-upgrade-testing
  kind delete cluster --name "${NAME}" -v9 || true
}

function setup_kind_cluster() {
  NAME=istio-upgrade-testing

  echo "Deleting previous KinD cluster with name=${NAME}"
  if ! (kind delete cluster --name="${NAME}" -v9) > /dev/null; then
    echo "No existing kind cluster with name ${NAME}. Continue..."
  fi

  # note: don't need private registry for upgrade tests, just
  # create kind cluster with default config
  kind create cluster --name ${NAME}
}

# metallb installation so we can access cluster through ingress service
# taken from istio/istio/prow scripts
cidr_to_ips() {
    CIDR="$1"
    python3 - <<EOF
from ipaddress import IPv4Network; [print(str(ip)) for ip in IPv4Network('$CIDR').hosts()]
EOF
}

metallb() {
  kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/namespace.yaml
  kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.9.3/manifests/metallb.yaml
  kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)" || true

  if [ -z "${METALLB_IPS[*]}" ]; then
    # Take IPs from the end of the docker kind network subnet to use for MetalLB IPs
    DOCKER_KIND_SUBNET="$(docker inspect kind | jq '.[0].IPAM.Config[0].Subnet' -r)"
    METALLB_IPS=()
    while read -r ip; do
      METALLB_IPS+=("$ip")
    done < <(cidr_to_ips "$DOCKER_KIND_SUBNET" | tail -n 100)
  fi

  # Give this cluster of those IPs
  RANGE="${METALLB_IPS[1]}-${METALLB_IPS[9]}"
  METALLB_IPS=("${METALLB_IPS[@]:10}")

  echo 'apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - '"$RANGE" | kubectl apply -f -
}
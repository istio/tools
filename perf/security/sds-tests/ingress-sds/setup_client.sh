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

function deploy_sleep() {
    local id="${1:?"please specify the gateway id"}"
    local ns="${2:?"please specify the namespace"}"
    local cs="${3:?"please specify the cluster"}"
    local wd="${4:?"please specify the cert path"}"
    local host="httpbin-${id}.example.com"
    local url="https://httpbin-${id}.example.com:$SECURE_INGRESS_PORT/status/418"

    # shellcheck disable=SC2154
    cat <<-EOF | kubectl apply -n "${ns}" --cluster "${cs}" -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sleep
---
apiVersion: v1
kind: Service
metadata:
  name: sleep
  labels:
    app: sleep
spec:
  ports:
  - port: 80
    name: http
  selector:
    app: sleep
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sleep
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sleep
  template:
    metadata:
      labels:
        app: sleep
    spec:
      serviceAccountName: sleep
      containers:
      - name: sleep
        image: istio/kubectl:1.3.0
        args:
          - bash
          - -c
          - |-
            sleep 60
            num_curl=0
            num_succeed=0

            while true; do
              resp_code=$(curl -sS  -o /dev/null -w "%{http_code}\n" -HHost:"${host}" --resolve "${host}":"${SECURE_INGRESS_PORT}":"${INGRESS_HOST}" --cacert "${wd}"/example.com.crt "${url}")
              if [ ${resp_code} = 200 ]; then
                num_succeed=$((num_succeed+1))
              else
                echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") curl failed, response code $resp_code"
              fi
              num_curl=$((num_curl+1))
              echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") Out of ${num_curl} curl, ${num_succeed} succeeded."
              sleep .5
            done
EOF
}
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
    local wd="${4:?"please specify the wd"}"
    local url="http:\/\/my-nginx-${id}.mesh-external.svc.cluster.local"

    sed "s/URL_TO_REPLACE/${url}/g" request_script.sh > "${wd}/request_script_${id}.sh"

    kubectl -n "${ns}" --cluster "${cs}" create configmap script-"${id}" --from-file="${wd}/request_script_${id}.sh"

    # shellcheck disable=SC2154
    cat <<-EOF | kubectl apply -n "${ns}" --cluster "${cs}" -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: "sleep-${id}"
---
apiVersion: v1
kind: Service
metadata:
  name: "sleep-${id}"
  labels:
    app: "sleep-${id}"
spec:
  ports:
  - port: 80
    name: http
  selector:
    app: "sleep-${id}"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: "sleep-${id}"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: "sleep-${id}"
  template:
    metadata:
      labels:
        app: "sleep-${id}"
    spec:
      serviceAccountName: "sleep-${id}"
      containers:
      - name: sleep
        image: istio/kubectl:1.3.0
        volumeMounts:
        - name: "script-${id}"
          mountPath: /opt/script
        args:
          - bash
          - -c
          - |-
            /bin/sh "/opt/script/request_script_${id}.sh"
      volumes:
        - name: "script-${id}"
          configMap:
            name: "script-${id}"
---
EOF
}

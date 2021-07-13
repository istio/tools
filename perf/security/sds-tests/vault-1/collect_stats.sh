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

NAMESPACE=${NAMESPACE:?"specify the namespace for running the test"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

echo "Run curl to test certificate rotations and mTLS"
num_curl=0
num_succeed=0


while true
do
  sleep_pods=$(kubectl get pods -n "${NAMESPACE}" -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -l app=sleep --cluster "${CLUSTER}")
  pods=()

  while read -r line; do
    pods+=("$line")
  done <<< "${sleep_pods}"

  if [ ${#pods[@]} = 0 ]; then
    echo "no pods found!"
  fi

  for pod in "${pods[@]}"
  do
    resp_code=$(kubectl exec -it -n "${NAMESPACE}" "${pod}" -c sleep --cluster "${CLUSTER}" -- curl -s -o /dev/null -w "%{http_code}" httpbin:8000/headers)
    if [ "${resp_code}" = 200 ]; then
      num_succeed=$((num_succeed+1))
    else
      echo "curl from the pod ${pod} failed"
    fi
    num_curl=$((num_curl+1))
    echo "Out of ${num_curl} curl, ${num_succeed} succeeded."
    sleep 1
  done

done
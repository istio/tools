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

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")

set -eux

declare -a namespaces=("istio-system" )
# shellcheck disable=SC2004
for ((ii=0; ii<15; ii++)) {
    ns=$(printf 'service-graph%.2d' ${ii})
    namespaces+=(${ns})
}

function check_events() {
  printf '\n'
  if [[ ${#ERRORED[@]} -ne 0 ]]
  then
      echo "${#ERRORED[@]} errored pods found."
      for CULPRIT in ${ERRORED[@]}
      do
        echo "POD: $CULPRIT"
        echo
        kubectl get events \
        --field-selector=involvedObject.name=${CULPRIT} \
        -ocustom-columns=LASTSEEN:.lastTimestamp,REASON:.reason,MESSAGE:.message \
        --all-namespaces \
        --ignore-not-found=true
      done
  else
      echo "0 pods with errored events found."
  fi
}

function check_pod_errors() {
  for NAMESPACE in ${namespaces[@]}
  do
    echo "Scanning pod logs, Namespace: ${NAMESPACE}"
    if ! kubectl get ns "${NAMESPACE}";then
      continue
    fi
    while IFS=' ' read -r POD CONTAINERS
    do
      for CONTAINER in ${CONTAINERS//,/ }
      do
        COUNT=$(kubectl logs --since=24h "${POD}" -c "${CONTAINER}" -n "${NAMESPACE}" | egrep -c '^error|Error|ERROR|Warn|WARN' || true)
        if [[ ${COUNT} -gt 0 ]];then
            STATE=("${STATE[@]}" "$POD|$CONTAINER|$COUNT")
        else
            ERRORED=($(kubectl get pods -n "${NAMESPACE}" --no-headers=true | \
                awk '!/Running/ {print $1}' ORS=" ") \
                )
        fi
      done
    done< <(kubectl get pods -n ${NAMESPACE} --ignore-not-found=true -o=custom-columns=NAME:.metadata.name,CONTAINERS:.spec.containers[*].name --no-headers=true)
  done
}

check_pod_errors
check_events
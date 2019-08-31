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

function redeploy() {
  local dlp=${1:?"deployment"}
  local namespace=${2:?"namespace"}
  kubectl patch deployment "${dpl}" \
      -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"`date +'%s'`\"}}}}}" \
      -n "${namespace}"
}


function redeploy_ns() {
  local namespace=${1:?"namespace"}
  for dpl in $(kubectl get deployments -o jsonpath="{.items[*].metadata.name}" -n ${namespace});do
    echo "Redeploy ${namespace}"
    redeploy "${dpl}" "${namespace}"
  done
}

function redeploy_all() {
  for ns in $(kubectl get ns -o jsonpath="{.items[*].metadata.name}" -listio-injection=enabled);do
    redeploy_ns "${ns}"
  done
}

function main() {
  local ns=${1:?" specific namespace or ALL"}

  if [[ "${ns}" == "ALL" ]];then
    redeploy_all
  else
    redeploy_ns "${ns}"
  fi
}

main "$*"

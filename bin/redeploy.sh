#!/bin/bash
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

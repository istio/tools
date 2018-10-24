#!/bin/bash

# This script spins up the standard 20 services per namespace test for as many namespaces
# as desired.
WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -ex

NUM=${1:?"number of namespaces. 20 x this number"}
START=${2:-"0"}

# service-graph04 svc04
CMD=""
if [[ ! -z "${ECHO}" ]];then
  CMD="echo" 
fi

function start_servicegraphs() {
  local nn=${1:?"number of namespaces"}
  local min=${2:-"0"}

  for ((ii=$min; ii<$nn; ii++)) {
    ns=$(printf 'service-graph%.2d' $ii)
    prefix=$(printf 'svc%.2d-' $ii)
    
    if [[ -z "${DELETE}" ]];then
      ${CMD} "${WD}/setup_test.sh" "${ns}" "${prefix}"
      ${CMD} "${WD}/../loadclient/setup_test.sh" "${ns}" "${prefix}"
    else
      ${CMD} "${WD}/../loadclient/setup_test.sh" "${ns}" "${prefix}"
      ${CMD} "${WD}/setup_test.sh" "${ns}" "${prefix}"
    fi    
  }
}

start_servicegraphs "${NUM}" "${START}"

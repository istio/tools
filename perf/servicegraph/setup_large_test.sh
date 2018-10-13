#!/bin/bash

# This script spins up the standard 20 services per namespace test for as many namespaces
# as desired.

WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)

NUM=${1:?"number of namespaces. 20 x this number"}

# service-graph04 svc04
CMD=""
if [[ ! -z "${ECHO}" ]];then
  CMD="echo" 
fi

function start_servicegraphs() {
  local nn=${1:?"number of namespaces"}

  for ((ii=0; ii<$nn; ii++)) {
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

start_servicegraphs "${NUM}"

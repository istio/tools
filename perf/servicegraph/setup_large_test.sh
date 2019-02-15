#!/bin/bash

# This script spins up the standard 20 services per namespace test for as many namespaces
# as desired.
WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -ex

source common.sh

NUM=${1:?"number of namespaces. 20 x this number"}
START=${2:-"0"}

# service-graph04 svc04
CMD=""
if [[ ! -z "${ECHO}" ]];then
  CMD="echo" 
fi

start_servicegraphs "${NUM}" "${START}"

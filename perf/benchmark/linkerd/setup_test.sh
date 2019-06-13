#!/bin/bash
WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -x
kubectl apply -f twopods-namespace.yaml 
kubectl apply -f fortio.yaml 
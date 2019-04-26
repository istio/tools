#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

gateway=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
port=80
address="$gateway:$port"
${WD}/../setup_test.sh "tcp-load" "--set address=$address"
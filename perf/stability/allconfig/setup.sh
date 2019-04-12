#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

gateway=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
domain=${DNS_DOMAIN:-qualistio.org}
${WD}/../setup_test.sh "allconfig" "--set ingress=${gateway} --set domain=${domain}"

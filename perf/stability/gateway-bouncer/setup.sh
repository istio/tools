#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

NAMESPACE="${NAMESPACE:-"istio-stability-gateway-bouncer"}"
${WD}/../setup_test.sh "gateway-bouncer" "--set namespace=${NAMESPACE}"

if [ -z ${DRY_RUN}; then
  # Waiting until LoadBalancer is created and retrieving the assigned
  # external IP address.
  while : ; do
    INGRESS_IP=$(kubectl -n ${NAMESPACE} \
      get service istio-ingress-${NAMESPACE} \
      -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

    if [[ -z "${INGRESS_IP}" ]]; then
      sleep 5s
    else
      break
    fi
  done

  # Populating a ConfigMap with the external IP address and restarting the
  # client to pick up the new version of the ConfigMap.
  kubectl -n ${NAMESPACE} delete configmap fortio-client-config
  kubectl -n ${NAMESPACE} create configmap fortio-client-config \
    --from-literal=external_addr=${INGRESS_IP}
  kubectl -n ${NAMESPACE} patch deployment fortio-client \
    -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"date\":\"`date +'%s'`\"}}}}}"
fi
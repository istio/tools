#!/bin/bash
set -ex

NAMESPACE=${NAMESPACE:?"New/existing NAMESPACE to install this scenario into"}

function install_gateway_bouncer() {
  local DIRNAME="${1:?"output dir"}"
  local OUTFILE="${DIRNAME}/gateway_bouncer.yaml"
  local INGRESS_IP=""

  helm -n $NAMESPACE template \
    --set namespace=$NAMESPACE \
    . > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
    kubectl -n $NAMESPACE apply -f "${OUTFILE}"

    # Waiting until LoadBalancer is created and retrieving the assigned
    # external IP address.
    while : ; do
      INGRESS_IP=$(kubectl -n $NAMESPACE \
        get service istio-ingress-$NAMESPACE \
        -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

      if [[ -z "${INGRESS_IP}" ]]; then
        sleep 5s
      else
        break
      fi
    done

    # Populating a ConfigMap with the external IP address and restarting the
    # client to pick up the new version of the ConfigMap.
    kubectl -n $NAMESPACE delete configmap fortio-client-config
    kubectl -n $NAMESPACE create configmap fortio-client-config \
      --from-literal=external_addr=$INGRESS_IP
    kubectl -n $NAMESPACE patch deployment fortio-client \
      -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"date\":\"`date +'%s'`\"}}}}}"
  fi
}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

install_gateway_bouncer "${WD}/tmp" $*

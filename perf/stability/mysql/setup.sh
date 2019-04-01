#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

source "${WD}/../common_setup.sh"

setup_test "mysql" "--set namespace=${NAMESPACE:-"mysql"}"
NAMESPACE="${NAMESPACE:-"mysql"}"

echo $WD

# verify mysql client can reach to server.
function verify_mysql() {
  kubectl -n ${NAMESPACE} exec ${client_pod} -c mysql-client \
-- mysql -uroot -proot -h mysql-server --connect-timeout 3 -P3306  -e 'show databases;'
}

echo "Wait till MySQL client and server are ready..."
client_pod=$(kubectl -n ${NAMESPACE} get pod -l app=mysql-client -o jsonpath='{.items[0].metadata.name}')
server_pod=$(kubectl -n ${NAMESPACE} get pod -l app=mysql-server -o jsonpath='{.items[0].metadata.name}')
kubectl -n${NAMESPACE} wait --for=condition=Ready pod/${client_pod}
kubectl -n${NAMESPACE} wait --for=condition=Ready pod/${server_pod}

# Disable mtls explicitly to avoid PERMISSIVE impact.
# TODO: add a link on istio.io.
echo "Testing MySQL mTLS is disabled, expect succeed..."
kubectl -n${NAMESPACE} apply -f mysql/mtls-disable.yaml
kubectl -n${NAMESPACE} apply -f mysql/mtls-enable.yaml -n ${NAMESPACE}
#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

source "${WD}/../common_setup.sh"

setup_test "mysql" "--set namespace=${NAMESPACE:-"mysql"}"
NAMESPACE="${NAMESPACE:-"mysql"}"

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
sleep 10
if verify_mysql; then
  echo "Succeed"
else
  echo "Failed"
fi

kubectl delete -f mysql/mtls-disable.yaml -n ${NAMESPACE}

echo "Testing MySQL mTLS is enabled, expect succeed..."
kubectl apply -f mysql/mtls-enable.yaml -n ${NAMESPACE}
sleep 10
if verify_mysql; then
  echo "Succeed"
else
  echo "Failed"
fi

echo "Testing MySQL mTLS is enabled, no destination rule, expect fail..."
kubectl delete dr mysql-mtls-dr -n ${NAMESPACE}
sleep 10
if verify_mysql; then
  echo "Failed"
else
  echo "Succeed"
fi

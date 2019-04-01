#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

source "${WD}/../common_setup.sh"

NAMESPACE="${NAMESPACE:-"mysql"}"
setup_test "mysql" "--set namespace=${NAMESPACE:-"mysql"} --set Name=mtls"
setup_test "mysql" "--set namespace=${NAMESPACE:-"mysql"} --set Name=plaintext"
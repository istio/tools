#!/bin/bash

set -ex

WD=$(dirname $0)
WD=$(cd $WD; pwd)

${WD}/../setup_test.sh "mysql" "--set Name=mtls"
${WD}/../setup_test.sh "mysql" "--set Name=plaintext"
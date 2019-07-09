#!/bin/bash

WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -ex

helm template . | kubectl -n istio-system apply -f -

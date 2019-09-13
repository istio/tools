#!/bin/bash

pushd ../../istio-install
export ISTIO_RELEASE=${ISTIO_RELEASE:-"release-1.2-latest"}
export DNS_DOMAIN=${DNS_DOMAIN:-"istio-automtls.local"}
export VALUES="values-auto-mtls.yaml"
./setup_istio.sh $ISTIO_RELEASE
popd


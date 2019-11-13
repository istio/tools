#!/bin/bash

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# pushd ../../istio-install || exit
# export ISTIO_RELEASE=${ISTIO_RELEASE:-"release-1.2-latest"}
# export DNS_DOMAIN=${DNS_DOMAIN:-"istio-automtls.local"}
# export VALUES="values-auto-mtls.yaml"
# ./setup_istio.sh "${ISTIO_RELEASE}"
# popd || exit

mkdir -p tmp
pushd tmp
URL=https://github.com/istio/istio/releases/download/1.4.0-beta.4/istio-1.4.0-beta.4-linux.tar.gz
wget $URL
tar xf istio-1.4.0-beta.4-linux.tar.gz

pushd istio-1.4.0-beta.4
bin/istioctl manifest apply --set profile=demo \
  --set values.global.mtls.auto=true \
  --set values.global.mtls.enabled=false
popd

popd
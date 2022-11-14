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

CONFIG_DIR=$(dirname "$0")

# Install ext-authz
kubectl apply -n twopods-istio -f https://raw.githubusercontent.com/istio/istio/release-1.15/samples/extauthz/ext-authz.yaml
kubectl patch configmap -n istio-system istio --patch-file "${CONFIG_DIR}/ext-authz_patch.yaml"
kubectl rollout restart deployment/istiod -n istio-system

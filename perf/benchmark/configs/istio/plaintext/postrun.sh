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

echo "Delete the plaintext related config..."
kubectl delete peerauthentication.security.istio.io -n"${NAMESPACE}" default
kubectl delete DestinationRule -n"${NAMESPACE}" plaintext-dr-twopods
echo "Restoring original Authn Policy and DestinationRule config..."

[[ -s "${LOCAL_OUTPUT_DIR}/authn-policy.yaml" ]] && kubectl apply -f "${LOCAL_OUTPUT_DIR}/authn-policy.yaml"
[[ -s "${LOCAL_OUTPUT_DIR}/destination-rule.yaml" ]] && kubectl apply -f "${LOCAL_OUTPUT_DIR}/destination-rule.yaml"

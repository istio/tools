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

echo "Saving current mTLS config first"
kubectl -n "${NAMESPACE}"  get dr -oyaml > "${LOCAL_OUTPUT_DIR}/destionation-rule.yaml" || true
kubectl -n "${NAMESPACE}"  get policy -oyaml > "${LOCAL_OUTPUT_DIR}/authn-policy.yaml" || true
echo "Deleting Authn Policy and DestinationRule"
kubectl -n "${NAMESPACE}" delete dr --all || true
kubectl -n "${NAMESPACE}" delete policy --all || true
echo "Configure plaintext..."
cat <<EOF | kubectl apply -f -
apiVersion: "authentication.istio.io/v1alpha1"
kind: "Policy"
metadata:
  name: "default"
  namespace: "${NAMESPACE}"
spec: {}
EOF
  # Explicitly disable mTLS by DestinationRule to avoid potential auto mTLS effect.
  cat <<EOF | kubectl apply -f -
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: plaintext-dr-twopods
  namespace: ${NAMESPACE}
spec:
  host:  "*.svc.${NAMESPACE}.cluster.local"
  trafficPolicy:
    tls:
      mode: DISABLE
EOF
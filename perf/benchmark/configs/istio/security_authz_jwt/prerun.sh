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

echo "Configure Security Policies..."
POLICY_PATH="${WD}/security/generate_policies"

echo "Build Security Policy Generator..."
go build -o "${LOCAL_OUTPUT_DIR}/generator" "${POLICY_PATH}/generate_policies.go" "${POLICY_PATH}/generate.go" "${POLICY_PATH}/jwt.go"

echo "Apply Security Policy to Cluster..."
./"${LOCAL_OUTPUT_DIR}/generator" -configFile="${CONFIG_DIR}/security_authz_jwt/config.json" > "${LOCAL_OUTPUT_DIR}/largeSecurityRequestAuthnPolicy.yaml"

cp "${CONFIG_DIR}/security_authz_jwt/latency.yaml" "${LOCAL_OUTPUT_DIR}/latency.yaml"

SECURITY_REQUEST_AUTHN_TOKEN=$(<token.txt)
echo "Generate Security Token for Requests:"
echo "$SECURITY_REQUEST_AUTHN_TOKEN"

envsubst < "${LOCAL_OUTPUT_DIR}/latency.yaml" > "${CONFIG_DIR}/security_request_authn/latency.yaml"

kubectl apply -f "${LOCAL_OUTPUT_DIR}/largeSecurityRequestAuthzJwtPolicy.yaml"

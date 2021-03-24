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
"${LOCAL_OUTPUT_DIR}/generator" -configFile="${CONFIG_DIR}/security_authz_ip/config.json" > "${LOCAL_OUTPUT_DIR}/largeSecurityAuthzIPPolicy.yaml"

kubectl apply -f "${LOCAL_OUTPUT_DIR}/largeSecurityAuthzIPPolicy.yaml"

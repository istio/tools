#!/usr/bin/env bash

# Copyright 2020 Istio Authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

WD=$(dirname "$0")
WD=$(cd "$WD" || exit; pwd)
# shellcheck disable=SC1090
source "${WD}/setup_security_test.sh"
# shellcheck disable=SC1090
source "${WD}/util/util.sh"

# Before running the security tests in this script:
# 1) The environmental variables in ./setup_security_test.sh must be configured
# based on your multi-cluster installation.
# 2) Login the project hosting your multi-cluster installation, e.g., through
# gcloud auth login
# gcloud config set project ${PROJECT_ID}

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
mkdir -p "${WD}/example-svc"
pushd "${WD}/example-svc"
pwd

# Download package that contains example service deployment files.
if [[ "$ISTIO_DOWNLOAD_METHOD" == "gsutil" ]]; then
    echo "Download $ISTIO_RELEASE_URL using gsutil"
    gsutil cp "$ISTIO_RELEASE_URL" .
elif [[ "$ISTIO_DOWNLOAD_METHOD" == "curl" ]]; then
    echo "Download $ISTIO_RELEASE_URL using curl"
    curl -LO "${ISTIO_RELEASE_URL}"
else
    echo "Exit due to invalid Istio download method: ${ISTIO_DOWNLOAD_METHOD}."
    exit 1
fi
tar xzf "${ISTIO_RELEASE_PKG}"
export ISTIO="${WD}"/example-svc/"${ISTIO_RELEASE_NAME}"
if [ -d "$ISTIO" ]; then
    echo "The Istio pkg is unzipped into ${ISTIO}."
else
    echo "Exit due to the error: the directory ${ISTIO} not found."
    exit 1
fi

if [[ -z "${PROJECT_ID}" || -z "${CLUSTER_1}" || -z "${CLUSTER_2}" || -z "${LOCATION_1}" || -z "${LOCATION_2}" ]]; then
    echo "Error: PROJECT_ID, CLUSTER_1, CLUSTER_2, LOCATION_1, LOCATION_2 must be set."
    exit 1
fi
export CTX_1=gke_${PROJECT_ID}_${LOCATION_1}_${CLUSTER_1}
export CTX_2=gke_${PROJECT_ID}_${LOCATION_2}_${CLUSTER_2}
gcloud container clusters get-credentials "${CLUSTER_1}" --zone "${LOCATION_1}" --project "${PROJECT_ID}"
gcloud container clusters get-credentials "${CLUSTER_2}" --zone "${LOCATION_2}" --project "${PROJECT_ID}"

# In cluster 1 and 2, create namespaces authz-ns1 and authz-ns2. In cluster 1 and 2, deploy helloworld and sleep services in namespaces authz-ns1 and authz-ns2.
kubectl create --context="${CTX_1}" namespace authz-ns1
kubectl label --context="${CTX_1}" namespace authz-ns1 \
  istio-injection=enabled
kubectl create --context="${CTX_1}" namespace authz-ns2
kubectl label --context="${CTX_1}" namespace authz-ns2 \
  istio-injection=enabled
kubectl create --context="${CTX_1}" \
  -f "${ISTIO}"/samples/helloworld/helloworld.yaml -n authz-ns1
kubectl create --context="${CTX_1}" \
  -f "${ISTIO}"/samples/helloworld/helloworld.yaml -n authz-ns2
kubectl apply --context="${CTX_1}" \
  -f "${ISTIO}"/samples/sleep/sleep.yaml -n authz-ns1
kubectl apply --context="${CTX_1}" \
  -f "${ISTIO}"/samples/sleep/sleep.yaml -n authz-ns2

kubectl create --context="${CTX_2}" namespace authz-ns1
kubectl label --context="${CTX_2}" namespace authz-ns1 \
  istio-injection=enabled
kubectl create --context="${CTX_2}" namespace authz-ns2
kubectl label --context="${CTX_2}" namespace authz-ns2 \
  istio-injection=enabled
kubectl create --context="${CTX_2}" \
  -f "${ISTIO}"/samples/helloworld/helloworld.yaml -n authz-ns1
kubectl create --context="${CTX_2}" \
  -f "${ISTIO}"/samples/helloworld/helloworld.yaml -n authz-ns2
kubectl apply --context="${CTX_2}" \
  -f "${ISTIO}"/samples/sleep/sleep.yaml -n authz-ns1
kubectl apply --context="${CTX_2}" \
  -f "${ISTIO}"/samples/sleep/sleep.yaml -n authz-ns2

# Verify the sleep and helloworld pods in authz-ns1 and authz-ns2
# of cluster 1 and 2 are ready.
waitForPodsInContextReady authz-ns1 "${CTX_1}" "2/2"
waitForPodsInContextReady authz-ns2 "${CTX_1}" "2/2"
waitForPodsInContextReady authz-ns1 "${CTX_2}" "2/2"
waitForPodsInContextReady authz-ns2 "${CTX_2}" "2/2"

# Verify sleep in authz-ns1 of cluster 1 can reach the helloworld
# service in authz-ns1 of both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_1}" -it -n authz-ns1 -c sleep \
  $(kubectl get pod --context="${CTX_1}" -n authz-ns1 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns1:5000/hello

# Verify sleep in authz-ns1 of cluster 1 can reach the helloworld
# service in authz-ns2 of both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_1}" -it -n authz-ns1 -c sleep \
  $(kubectl get pod --context="${CTX_1}" -n authz-ns1 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns2:5000/hello

# Verify sleep in authz-ns2 of cluster 1 can reach the helloworld
# service in authz-ns1 of both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_1}" -it -n authz-ns2 -c sleep \
  $(kubectl get pod --context="${CTX_1}" -n authz-ns2 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns1:5000/hello

# Verify sleep in authz-ns2 of cluster 1 can reach the helloworld
# service in authz-ns2 of both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_1}" -it -n authz-ns2 -c sleep \
  $(kubectl get pod --context="${CTX_1}" -n authz-ns2 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns2:5000/hello

# Verify sleep in authz-ns1 of cluster 2 can reach the helloworld
# service in authz-ns1 of both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_2}" -it -n authz-ns1 -c sleep \
  $(kubectl get pod --context="${CTX_2}" -n authz-ns1 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns1:5000/hello

# Verify sleep in authz-ns1 of cluster 2 can reach the helloworld
# service in authz-ns2 of both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_2}" -it -n authz-ns1 -c sleep \
  $(kubectl get pod --context="${CTX_2}" -n authz-ns1 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns2:5000/hello

# Verify sleep in authz-ns2 of cluster 2 can reach the helloworld
# service in authz-ns1 of both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_2}" -it -n authz-ns2 -c sleep \
  $(kubectl get pod --context="${CTX_2}" -n authz-ns2 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns1:5000/hello

# Verify sleep in authz-ns2 of cluster 2 can reach the helloworld
# service in authz-ns2 in both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_2}" -it -n authz-ns2 -c sleep \
  $(kubectl get pod --context="${CTX_2}" -n authz-ns2 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns2:5000/hello

# Deploy an authorization policy in cluster 1 and 2 to deny traffic to authz-ns2.
kubectl apply --context="${CTX_1}" -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: deny-authz-ns2
  namespace: authz-ns2
spec:
  {}
EOF

kubectl apply --context="${CTX_2}" -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: deny-authz-ns2
  namespace: authz-ns2
spec:
  {}
EOF

# Wait 60 seconds for the policies to take effect.
echo "Wait 60 seconds for the policies to take effect."
sleep 60

# Verify traffic from sleep in authz-ns1 of cluster 1 to helloworld.authz-ns2 is denied.
verifyResponses 5 0 "RBAC: access denied" kubectl exec --context="${CTX_1}" -it -n authz-ns1 -c sleep \
  $(kubectl get pod --context="${CTX_1}" -n authz-ns1 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns2:5000/hello

# Verify traffic from sleep in authz-ns1 of cluster 2 to helloworld.authz-ns2 is denied.
verifyResponses 5 0 "RBAC: access denied" kubectl exec --context="${CTX_2}" -it -n authz-ns1 -c sleep \
  $(kubectl get pod --context="${CTX_2}" -n authz-ns1 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns2:5000/hello

# Verify traffic from sleep in authz-ns2 of cluster 2 is allowed by
# all helloworld service instances in authz-ns1 of both clusters.
# The response set should include those from 4 instances in all clusters.
verifyResponseSet 10 0 4 kubectl exec --context="${CTX_2}" -it -n authz-ns2 -c sleep \
  $(kubectl get pod --context="${CTX_2}" -n authz-ns2 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns1:5000/hello

# Deploy an authorization policy in cluster 1 and 2 to deny traffic to authz-ns1.
kubectl apply --context="${CTX_1}" -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: deny-authz-ns1
  namespace: authz-ns1
spec:
  {}
EOF

kubectl apply --context="${CTX_2}" -f - <<EOF
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: deny-authz-ns1
  namespace: authz-ns1
spec:
  {}
EOF

# Wait 60 seconds for the policies to take effect.
echo "Wait 60 seconds for the policies to take effect."
sleep 60

# Verify traffic from sleep in authz-ns2 of cluster 2 to helloworld.authz-ns1 is denied.
verifyResponses 5 0 "RBAC: access denied" kubectl exec --context="${CTX_2}" -it -n authz-ns2 -c sleep \
  $(kubectl get pod --context="${CTX_2}" -n authz-ns2 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns1:5000/hello

# Verify traffic from sleep in authz-ns1 of cluster 1 to helloworld.authz-ns1 is denied.
verifyResponses 5 0 "RBAC: access denied" kubectl exec --context="${CTX_1}" -it -n authz-ns1 -c sleep \
  $(kubectl get pod --context="${CTX_1}" -n authz-ns1 -l \
  app=sleep -o jsonpath='{.items[0].metadata.name}') -- curl \
  helloworld.authz-ns1:5000/hello

# Clean up the resources created after the tests
popd
# shellcheck disable=SC1090
source "${WD}/cleanup_authz_security_tests.sh"

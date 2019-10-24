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

NAMESPACE=${NAMESPACE:?"specify the namespace for running the test"}
NUM=${NUM:?"specify the number of httpbin and sleep workloads"}
RELEASE=${RELEASE:?"specify the Istio release, e.g., release-1.1-20190208-09-16"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

# Download the istioctl
WD=$(dirname "$0")/tmp
if [[ ! -d "${WD}" ]]; then
  mkdir "$WD"
fi
wget -O "$WD/istio-${RELEASE}-linux.tar.gz" "https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${RELEASE}/istio-${RELEASE}-linux.tar.gz"
tar xfz "${WD}"/istio-"${RELEASE}"-linux.tar.gz -C "$WD"

function inject_workload() {
  local deployfile="${1:?"please specify the workload deployment file"}"
  # This test uses perf/istio/values-istio-sds-vault.yaml, in which
  # Istio auto sidecar injector is not enabled.
  "$WD"/istio-"${RELEASE}"/bin/istioctl kube-inject -f "${deployfile}" -o temp-workload-injected.yaml
  kubectl apply -n "${NAMESPACE}" -f temp-workload-injected.yaml --cluster "${CLUSTER}"
}

TEMP_DEPLOY_NAME="temp_httpbin_sleep_deploy.yaml"
helm template --set replicas="${NUM}" . > "${TEMP_DEPLOY_NAME}"

kubectl create ns "${NAMESPACE}" --cluster "${CLUSTER}"

kubectl create serviceaccount vault-citadel-sa -n "${NAMESPACE}"

SA_SECRET_NAME=$(kubectl get serviceaccount vault-citadel-sa -n "${NAMESPACE}" -o=jsonpath='{.secrets[0].name}')
export SA_SECRET_NAME

kubectl patch secret "${SA_SECRET_NAME}" -n "${NAMESPACE}" -p='{"data":{"token": "ZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNklpSjkuZXlKcGMzTWlPaUpyZFdKbGNtNWxkR1Z6TDNObGNuWnBZMlZoWTJOdmRXNTBJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5dVlXMWxjM0JoWTJVaU9pSmtaV1poZFd4MElpd2lhM1ZpWlhKdVpYUmxjeTVwYnk5elpYSjJhV05sWVdOamIzVnVkQzl6WldOeVpYUXVibUZ0WlNJNkluWmhkV3gwTFdOcGRHRmtaV3d0YzJFdGRHOXJaVzR0TnpSMGQzTWlMQ0pyZFdKbGNtNWxkR1Z6TG1sdkwzTmxjblpwWTJWaFkyTnZkVzUwTDNObGNuWnBZMlV0WVdOamIzVnVkQzV1WVcxbElqb2lkbUYxYkhRdFkybDBZV1JsYkMxellTSXNJbXQxWW1WeWJtVjBaWE11YVc4dmMyVnlkbWxqWldGalkyOTFiblF2YzJWeWRtbGpaUzFoWTJOdmRXNTBMblZwWkNJNklqSmhZekF6WW1FeUxUWTVNVFV0TVRGbE9TMDVOamt3TFRReU1ERXdZVGhoTURFeE5DSXNJbk4xWWlJNkluTjVjM1JsYlRwelpYSjJhV05sWVdOamIzVnVkRHBrWldaaGRXeDBPblpoZFd4MExXTnBkR0ZrWld3dGMyRWlmUS5wWjhTaXlOZU8wcDFwOEhCOW9YdlhPQUkxWENKWktrMndWSFhCc1RTektXeGxWRDlIckhiQWNTYk8yZGxoRnBlQ2drbnQ2ZVp5d3ZoU2haSmgyRjYtaUhQX1lvVVZvQ3FRbXpqUG9CM2MzSm9ZRnBKby05alROMV9tTlJ0WlVjTnZZbC10RGxUbUJsYUtFdm9DNVAyV0dWVUYzQW9Mc0VTNjZ1NEZHOVdsbG1MVjkyTEcxV05xeF9sdGtUMXRhaFN5OVdpSFFneXpQcXd0d0U3MlQxakFHZGdWSW9KeTFsZlNhTGFtX2JvOXJxa1JsZ1NnLWF1OUJBalppREd0bTl0ZjNsd3JjZ2ZieGNjZGxHNGpBc1RGYTJhTnMzZFc0TkxrN21GbldDSmEtaVdqLVRnRnhmOVRXLTlYUEswZzNvWUlRMElkMENJVzJTaUZ4S0dQQWpCLWc="}}'

inject_workload ${TEMP_DEPLOY_NAME}

echo "Wait 90 seconds for the deployment to be ready ..."
sleep 90

# shellcheck disable=SC1091
source ./collect_stats.sh

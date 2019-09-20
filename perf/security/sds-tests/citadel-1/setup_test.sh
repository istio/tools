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
RELEASETYPE=${RELEASETYPE:?"specify the Istio release type, daily, release, pre-release"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

# Download the istioctl
# shellcheck disable=SC2086
WD=$(dirname $0)/tmp
if [[ ! -d "${WD}" ]]; then
  # shellcheck disable=SC2086
  mkdir $WD
fi

# shellcheck disable=SC1091
source ../../utils/get_release.sh
# shellcheck disable=SC2086
get_release_url $RELEASETYPE $RELEASE
# shellcheck disable=SC2154
if [[ -z "$release_url" ]]; then
  exit
fi

curl -JLo "$WD/istio-${RELEASE}.tar.gz" "${release_url}"
# shellcheck disable=SC2086
tar xfz ${WD}/istio-${RELEASE}.tar.gz -C $WD

function inject_workload() {
  local deployfile="${1:?"please specify the workload deployment file"}"
  # This test uses perf/istio/values-istio-sds-auth.yaml, in which
  # Istio auto sidecar injector is not enabled.
  # shellcheck disable=SC2086
  $WD/istio-${RELEASE}/bin/istioctl kube-inject -f "${deployfile}" -o temp-workload-injected.yaml
  # shellcheck disable=SC2086
  kubectl apply -n ${NAMESPACE} -f temp-workload-injected.yaml --cluster ${CLUSTER}
}

TEMP_DEPLOY_NAME="temp_httpbin_sleep_deploy.yaml"
helm template --set replicas="${NUM}" ../../workload-deployments/ > "${TEMP_DEPLOY_NAME}"

# shellcheck disable=SC2086
kubectl create ns ${NAMESPACE} --cluster ${CLUSTER}

inject_workload ${TEMP_DEPLOY_NAME}

echo "Wait 60 seconds for the deployment to be ready ..."
sleep 60

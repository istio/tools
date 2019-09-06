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
  local tempdeployfile="${1:?"please specify the template workload deployment file"}"
  local deployfile="${2:?"please specify the workload deployment file"}"
  # This test uses perf/istio/values-istio-sds-auth.yaml, in which
  # Istio auto sidecar injector is not enabled.
  # shellcheck disable=SC2086
  $WD/istio-${RELEASE}/bin/istioctl kube-inject -f "${tempdeployfile}" -o "${deployfile}"
  # shellcheck disable=SC2086
  kubectl apply -n ${NAMESPACE} -f "${deployfile}" --cluster ${CLUSTER}

  echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") Wait 180 seconds to generate some traffic ..."
  sleep 180
}

function rotate_deployment() {
  local deployfile="${1:?"please specify the workload deployment file"}"
  local workloaddeploymentconfig="workload-deploy"
  local workloadlife=180
  ROTATE_WORKLOAD_YAML="rotate_workload_deploy.yaml"

  # shellcheck disable=SC2086
  helm -n ${NAMESPACE} template \
    --set Namespace="${NAMESPACE}" \
    --set DeployYaml="${deployfile}" \
    --set WorkloadLife="${workloadlife}" \
    --set ConfigName="${workloaddeploymentconfig}" \
          . > "${ROTATE_WORKLOAD_YAML}"
  echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") Wrote ${ROTATE_WORKLOAD_YAML}"
  # Create ConfigMap workload-deploy and load workload deployment file into ConfigMap workload-deploy.
  echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") kubectl -n ${NAMESPACE} create configmap ${workloaddeploymentconfig} --from-file=${deployfile}=${deployfile}"
  kubectl -n "${NAMESPACE}" create configmap "${workloaddeploymentconfig}" --from-file="${deployfile}"="${deployfile}" --cluster "${CLUSTER}"
  # Create ConfigMap script and load workload rotate script into ConfigMap script.
  # Create a deployment to mount the workload deployment file and rotate script, and execute the script
  # to rotate workload deployment periodically.
  echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") kubectl apply -n ${NAMESPACE} -f ${ROTATE_WORKLOAD_YAML} --cluster ${CLUSTER}"
  # shellcheck disable=SC2086
  kubectl apply -n ${NAMESPACE} -f ${ROTATE_WORKLOAD_YAML} --cluster ${CLUSTER}
}

# shellcheck disable=SC2086
kubectl create ns ${NAMESPACE} --cluster ${CLUSTER}

TEMP_DEPLOY_FILE="temp_httpbin_sleep_deploy.yaml"
helm template --set replicas="${NUM}" ../../workload-deployments/ > "${TEMP_DEPLOY_FILE}"

DEPLOY_FILE="workload-injected.yaml"

inject_workload ${TEMP_DEPLOY_FILE} ${DEPLOY_FILE}

rotate_deployment ${DEPLOY_FILE}


#!/bin/bash

# Copyright 2019 Istio Authors
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

MASON_CLIENT_PID=-1
KUBE_USER="${KUBE_USER:-istio-prow-test-job@istio-testing.iam.gserviceaccount.com}"
SETUP_CLUSTERREG="${SETUP_CLUSTERREG:-False}"
USE_GKE="${USE_GKE:-True}"
SA_NAMESPACE="istio-system-multi"

# Set up a GKE cluster for testing.
function setup_e2e_cluster() {
  WD=$(dirname "$0")
  WD=$(cd "$WD" || exit; pwd)

  trap cleanup EXIT

  if [[ "${USE_MASON_RESOURCE}" == "True" ]]; then
    INFO_PATH="$(mktemp /tmp/XXXXX.boskos.info)"
    FILE_LOG="$(mktemp /tmp/XXXXX.boskos.log)"
    OWNER=${OWNER:-"e2e"}
    E2E_ARGS+=("--mason_info=${INFO_PATH}")

    setup_and_export_git_sha

    get_resource "${RESOURCE_TYPE}" "${OWNER}" "${INFO_PATH}" "${FILE_LOG}"
  else
    export GIT_SHA="${GIT_SHA:-$TAG}"
  fi
  setup_cluster
}

# Cleanup e2e resources.
function cleanup() {
  if [[ "${CLEAN_CLUSTERS}" == "True" ]]; then
    unsetup_clusters
  fi
  if [[ "${USE_MASON_RESOURCE}" == "True" ]]; then
    mason_cleanup
    cat "${FILE_LOG}"
  fi
}

function mason_cleanup() {
  if [[ ${MASON_CLIENT_PID} != -1 ]]; then
    kill -SIGINT ${MASON_CLIENT_PID} || echo "failed to kill mason client"
    wait
  fi
}

# This function would get sha from tools repo instead of istio releases.
function setup_and_export_git_sha() {
  if [[ -n "${CI:-}" ]]; then
    if [ -z "${PULL_PULL_SHA:-}" ]; then
      if [ -z "${PULL_BASE_SHA:-}" ]; then
        GIT_SHA="$(git rev-parse --verify HEAD)"
        export GIT_SHA
      else
        export GIT_SHA="${PULL_BASE_SHA}"
      fi
    else
      export GIT_SHA="${PULL_PULL_SHA}"
    fi
  else
    # Use the current commit.
    GIT_SHA="$(git rev-parse --verify HEAD)"
    export GIT_SHA
    export ARTIFACTS="${ARTIFACTS:-$(mktemp -d)}"
  fi
  GIT_BRANCH="$(git rev-parse --abbrev-ref HEAD)"
  export GIT_BRANCH
  setup_gcloud_credentials
}

function setup_gcloud_credentials() {
  if [[ $(command -v gcloud) ]]; then
    gcloud auth configure-docker -q
  elif [[ $(command -v docker-credential-gcr) ]]; then
    docker-credential-gcr configure-docker
  else
    echo "No credential helpers found, push to docker may not function properly"
  fi
}

# use mason_client to get resources
function get_resource() {
  # cd to tmp, otherwise the go.mod file may be updated
  # shellcheck disable=SC2164
  local type="${1}"
  local owner="${2}"
  local info_path="${3}"
  local file_log="${4}"

  mason_client \
    --type="${type}" \
    --boskos-url='http://boskos.boskos.svc.cluster.local' \
    --owner="${owner}" \
    --info-save "${info_path}" \
    --kubeconfig-save "${HOME}/.kube/config" > "${file_log}" 2>&1 &
  MASON_CLIENT_PID=$!

  local ready
  local exited

  # Wait up to 10 mn by increment of 10 seconds unit ready or failure
  for _ in {1..60}; do
    grep -q READY "${file_log}" && ready=true || ready=false
    if [[ ${ready} == true ]]; then
      cat "${info_path}"
      local project
      project="$(head -n 1 "${info_path}" | tr -d ':')"
      gcloud config set project "${project}"
      return 0
    fi
    kill -s 0 ${MASON_CLIENT_PID} && exited=false || exited=true
    if [[ ${exited} == true ]]; then
      cat "${file_log}"
      echo "Failed to get a Boskos resource"
      return 1
    fi
    sleep 10
  done
  echo 'failed to get a Boskos resource'
  return 1
}

function join_by { local IFS="$1"; shift; echo "$*"; }

function join_lines_by_comma() {
  # Turn each line into an element in an array.
  mapfile -t array <<< "$1"
  list=$(join_by , "${array[@]}")
  echo "${list}"
}

function setup_cluster() {
  # use current-context if pilot_cluster not set
  PILOT_CLUSTER="${PILOT_CLUSTER:-$(kubectl config current-context)}"

  unset IFS
  k_contexts=$(kubectl config get-contexts -o name)
  for context in ${k_contexts}; do
     kubectl config use-context "${context}"

     kubectl create clusterrolebinding prow-cluster-admin-binding\
       --clusterrole=cluster-admin\
       --user="${KUBE_USER}"
  done
  if [[ "${SETUP_CLUSTERREG}" == "True" ]]; then
      setup_clusterreg
  fi
  kubectl config use-context "${PILOT_CLUSTER}"

  if [[ "${USE_GKE}" == "True" && "${SETUP_CLUSTERREG}" == "True" ]]; then
    echo "Set up firewall rules."
    date
    ALL_CLUSTER_CIDRS_LINES=$(gcloud container clusters list --format='value(clusterIpv4Cidr)' | sort | uniq)
    ALL_CLUSTER_CIDRS=$(join_lines_by_comma "${ALL_CLUSTER_CIDRS_LINES}")

    ALL_CLUSTER_NETTAGS_LINES=$(gcloud compute instances list --format='value(tags.items.[0])' | sort | uniq)
    ALL_CLUSTER_NETTAGS=$(join_lines_by_comma "${ALL_CLUSTER_NETTAGS_LINES}")

    gcloud compute firewall-rules create istio-multicluster-test-pods \
	    --allow=tcp,udp,icmp,esp,ah,sctp \
	    --direction=INGRESS \
	    --priority=900 \
	    --source-ranges="${ALL_CLUSTER_CIDRS}" \
	    --target-tags="${ALL_CLUSTER_NETTAGS}" --quiet
  fi
}

function unsetup_clusters() {
  # use current-context if pilot_cluster not set
  PILOT_CLUSTER="${PILOT_CLUSTER:-$(kubectl config current-context)}"

  unset IFS
  k_contexts=$(kubectl config get-contexts -o name)
  for context in ${k_contexts}; do
     kubectl config use-context "${context}"

     kubectl delete clusterrolebinding prow-cluster-admin-binding 2>/dev/null
     if [[ "${SETUP_CLUSTERREG}" == "True" && "${PILOT_CLUSTER}" != "$context" ]]; then
        kubectl delete clusterrolebinding istio-multi-test 2>/dev/null
        kubectl delete ns ${SA_NAMESPACE} 2>/dev/null
     fi
  done
  kubectl config use-context "${PILOT_CLUSTER}"
  if [[ "${USE_GKE}" == "True" && "${SETUP_CLUSTERREG}" == "True" ]]; then
     gcloud compute firewall-rules delete istio-multicluster-test-pods --quiet
  fi
}

# setup_cluster_reg is used to set up a cluster registry for multicluster testing
function setup_cluster_reg () {
    MAIN_CONFIG=""
    for context in "${CLUSTERREG_DIR}"/*; do
        if [[ -z "${MAIN_CONFIG}" ]]; then
            MAIN_CONFIG="${context}"
        fi
        export KUBECONFIG="${context}"
        kubectl delete ns istio-system-multi --ignore-not-found
        kubectl delete clusterrolebinding istio-multi-test --ignore-not-found
        kubectl create ns istio-system-multi
        kubectl create sa istio-multi-test -n istio-system-multi
        kubectl create clusterrolebinding istio-multi-test --clusterrole=cluster-admin --serviceaccount=istio-system-multi:istio-multi-test
        CLUSTER_NAME=$(kubectl config view --minify=true -o "jsonpath={.clusters[].name}")
        gen_kubeconf_from_sa istio-multi-test "${context}"
    done
    export KUBECONFIG="${MAIN_CONFIG}"
}

function gen_kubeconf_from_sa () {
    local service_account=$1
    local filename=$2

    SERVER=$(kubectl config view --minify=true -o "jsonpath={.clusters[].cluster.server}")
    SECRET_NAME=$(kubectl get sa "${service_account}" -n istio-system-multi -o jsonpath='{.secrets[].name}')
    CA_DATA=$(kubectl get secret "${SECRET_NAME}" -n istio-system-multi -o "jsonpath={.data['ca\\.crt']}")
    TOKEN=$(kubectl get secret "${SECRET_NAME}" -n istio-system-multi -o "jsonpath={.data['token']}" | base64 --decode)

    cat <<EOF > "${filename}"
      apiVersion: v1
      clusters:
         - cluster:
             certificate-authority-data: ${CA_DATA}
             server: ${SERVER}
           name: ${CLUSTER_NAME}
      contexts:
         - context:
             cluster: ${CLUSTER_NAME}
             user: ${CLUSTER_NAME}
           name: ${CLUSTER_NAME}
      current-context: ${CLUSTER_NAME}
      kind: Config
      preferences: {}
      users:
         - name: ${CLUSTER_NAME}
           user:
             token: ${TOKEN}
EOF
}

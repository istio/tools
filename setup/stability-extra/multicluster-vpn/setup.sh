#!/usr/bin/env bash

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

# Configures an 'external cluster' for the `multicluster-vpn` scenarios.
# See `README.md` for details.
#
# This script is configured via the following set of environment variables (most of the arguments have a reasonable default value and can be left out):
#
# HUB                  -- Istio container images repository
# TAG                  -- Istio container images version tag
# APPS_NAMESPACE       -- Should be the same as the corresponding namespace in the primary stability cluster
# ISTIOCTL             -- Custom path to `istioctl`
# ISTIO_INSTALLER_PATH -- Custom path to Istio Installer (will be downloaded if not provided)
# KUBECONFIG1          -- `kubeconfig` of the primary stability cluster
# KUBECONTEXT1         -- context of the primary stability cluster
# KUBECONFIG2          -- `kubeconfig` of the external cluster that's being configured
# KUBECONTEXT2         -- context of the external cluster that's being configured
#
# A complete example:
#
# ```
# $ ISTIO_INSTALLER_PATH=~/works/installer \
#   ISTIOCTL="go run istio.io/istio/istioctl/cmd/istioctl" \
#   HUB=gcr.io/istio-release \
#   TAG=master-latest-daily \
#   KUBECONTEXT1=gke_istio-test-230101_us-central1-a_mc-s \
#   KUBECONTEXT2=gke_istio-test-230101_us-west1-a_mc-uswest \
#   ./setup.sh
# ```

set -e

HUB="${HUB:-gcr.io/istio-release}"
TAG="${TAG:-master-latest-daily}"
APPS_NAMESPACE="${APPS_NAMESPACE:-istio-stability-multicluster-vpn}"
ISTIOCTL="${ISTIOCTL:-istioctl}"

# KUBECONFIG: path to a kubeconfig file
KUBECONFIG1="${KUBECONFIG1:-${HOME}/.kube/config}"
KUBECONFIG2="${KUBECONFIG2:-${HOME}/.kube/config}"
# KUBECONTEXT: empty value defaults to "current" context of given kubeconfig file
KUBECONTEXT1="${KUBECONTEXT1:-$(kubectl --kubeconfig=${KUBECONFIG1} config current-context)}"
KUBECONTEXT2="${KUBECONTEXT2:-$(kubectl --kubeconfig=${KUBECONFIG2} config current-context)}"

### simple sanity check
if [[ $KUBECONFIG1 == $KUBECONFIG2 ]] && [[ $KUBECONTEXT1 == $KUBECONTEXT2 ]]; then
  echo
  echo " [FAIL] KUBECONFIG{1,2}/KUBECONTEXT{1,2} pairs refer to the same cluster"
  echo "        this configuration requires two distinct clusters, terminating..."
  echo
  exit 1
fi
###

BASE_DIR=$(dirname "$0")
TEMP_DIR=$(mktemp -d)

if [[ -z "$ISTIO_INSTALLER_PATH" ]]; then
  echo
  echo " [*] Istio Installer path was not provided. Cloning it into a temp location."
  echo

  ISTIO_INSTALLER_PATH="${TEMP_DIR}/installer"
  git clone https://github.com/istio/installer.git $ISTIO_INSTALLER_PATH
fi

function copy_istio_secrets {
  local KUBECTL_SRC="${1:?required argument is not set or empty}"
  local KUBECTL_DST="${2:?required argument is not set or empty}"

  $KUBECTL_DST -n istio-system scale deployment istio-citadel --replicas=0 || true
  $KUBECTL_DST -n istio-system rollout status deployment istio-citadel || true

  $KUBECTL_DST -n istio-system delete secret istio-ca-secret || true
  $KUBECTL_DST -n istio-system delete secret cacerts || true

  for ns in `$KUBECTL_DST get ns -o=jsonpath="{.items[*].metadata.name}"`; do
    echo $ns
    $KUBECTL_DST -n $ns delete secret istio.default || true
  done

  PLUGGED_SECRET=$($KUBECTL_SRC -n istio-system get secret cacerts -o yaml --export || true)
  if [[ -n "$PLUGGED_SECRET" ]]; then
    echo "$PLUGGED_SECRET" | $KUBECTL_DST -n istio-system apply --validate=false -f -
  fi

  SELFSIGNED_SECRET=$($KUBECTL_SRC -n istio-system get secret istio-ca-secret -o yaml --export || true)
  if [[ -n "$SELFSIGNED_SECRET" ]]; then
    echo "$SELFSIGNED_SECRET" | $KUBECTL_DST -n istio-system apply --validate=false -f -
  fi

  $KUBECTL_DST -n istio-system scale deployment istio-citadel --replicas=1 || true
  $KUBECTL_DST -n istio-system rollout status deployment istio-citadel || true

  sleep 5
}

function install_k8s_secrets {
  local KUBECONFIG_MASTER="${1:?required argument is not set or empty}"
  local KUBECONTEXT_MASTER="${2:?required argument is not set or empty}"
  local NAMESPACE_MASTER="${3:?required argument is not set or empty}"
  local KUBECONFIG_SLAVE="${4:?required argument is not set or empty}"
  local KUBECONTEXT_SLAVE="${5:?required argument is not set or empty}"
  local NAMESPACE_SLAVE="${6:?required argument is not set or empty}"

  local KUBECTL_MASTER="kubectl --kubeconfig=${KUBECONFIG_MASTER} --context=${KUBECONTEXT_MASTER}"
  local KUBECTL_SLAVE="kubectl --kubeconfig=${KUBECONFIG_SLAVE} --context=${KUBECONTEXT_SLAVE}"

  local CLUSTER_NAME=$($KUBECTL_SLAVE config view -o jsonpath="{.contexts[?(@.name == \"${KUBECONTEXT_SLAVE}\")].context.cluster}")
  local SERVER=$($KUBECTL_SLAVE config view -o jsonpath="{.clusters[?(@.name == \"${CLUSTER_NAME}\")].cluster.server}")
  local SERVICE_ACCOUNT=istio-pilot-service-account
  local SECRET_NAME=$($KUBECTL_SLAVE get sa ${SERVICE_ACCOUNT} -n ${NAMESPACE_SLAVE} -o jsonpath="{.secrets[].name}")
  local CA_DATA=$($KUBECTL_SLAVE get secret ${SECRET_NAME} -n ${NAMESPACE_SLAVE} -o jsonpath="{.data['ca\.crt']}")

  local TOKEN=$($KUBECTL_SLAVE get secret ${SECRET_NAME} -n ${NAMESPACE_SLAVE} -o jsonpath="{.data['token']}" | base64 --decode)

  local KUBECFG_FILE=$TEMP_DIR/kubeconfig
  cat > $KUBECFG_FILE <<EOF
apiVersion: v1
kind: Config
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
preferences: {}
users:
- name: ${CLUSTER_NAME}
  user:
    token: ${TOKEN}
EOF

  local SECRET_NAME=$(cat /dev/urandom | env LC_CTYPE=C tr -dc 'a-z0-9' | fold -w 32 | head -n 1)

  $KUBECTL_MASTER create secret generic ${SECRET_NAME} --from-file ${KUBECFG_FILE} -n ${NAMESPACE_MASTER}
  $KUBECTL_MASTER label secret ${SECRET_NAME} istio/multiCluster=true -n ${NAMESPACE_MASTER}

  sleep 5
}


KUBECTL1="kubectl --kubeconfig=${KUBECONFIG1} --context=${KUBECONTEXT1}"
KUBECTL2="kubectl --kubeconfig=${KUBECONFIG2} --context=${KUBECONTEXT2}"

echo
echo " [*] Installing Istio CRDs"
echo
$KUBECTL2 apply -f $ISTIO_INSTALLER_PATH/crds/files

echo
echo " [*] Copying Citadel secrets from the primary cluster and installing Citadel singleton"
echo
$KUBECTL2 create ns istio-system || true
copy_istio_secrets "$KUBECTL1" "$KUBECTL2"
helm template \
  --namespace istio-system \
  -n citadel \
  $ISTIO_INSTALLER_PATH/security/citadel/ \
  -f $ISTIO_INSTALLER_PATH/global.yaml \
  --set global.hub=$HUB \
  --set global.tag=$TAG \
  | $KUBECTL2 apply -f -

echo
echo " [*] Installing test scenarios"
echo
$KUBECTL2 create ns $APPS_NAMESPACE || true

( . $BASE_DIR/default/setup.sh )
( . $BASE_DIR/locality-distribute/setup.sh )
( . $BASE_DIR/locality-failover/setup.sh )

echo
echo " [*] Done"

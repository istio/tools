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

set -e

echo
echo " [*] Installing the 'default' scenario"
echo

SCENARIO="default"
NAMESPACE="env-$SCENARIO"
$KUBECTL2 create ns $NAMESPACE || true

$KUBECTL2 -n $NAMESPACE delete secrets -l "istio/multiCluster=true"
install_k8s_secrets "$KUBECONFIG2" "$KUBECONTEXT2" "$NAMESPACE" "$KUBECONFIG1" "$KUBECONTEXT1" "istio-system"

# shellcheck disable=SC2086
cat > ${TEMP_DIR}/values-${SCENARIO}.yaml <<EOF
global:
  hub: ${HUB}
  tag: ${TAG}
  istioNamespace: ${NAMESPACE}
  configNamespace: ${NAMESPACE}
  telemetryNamespace: ${NAMESPACE}
  policyNamespace: ${NAMESPACE}
  configValidation: false
  proxy:
    tracer: ""
pilot:
  configNamespace: ${NAMESPACE}
EOF

# shellcheck disable=SC2086
helm template \
  --namespace $NAMESPACE \
  -n galley-$NAMESPACE \
  $ISTIO_INSTALLER_PATH/istio-control/istio-config/ \
  -f $ISTIO_INSTALLER_PATH/global.yaml \
  -f ${TEMP_DIR}/values-${SCENARIO}.yaml \
  | $KUBECTL2 apply -f -

# shellcheck disable=SC2086
helm template \
  --namespace $NAMESPACE \
  -n pilot-$NAMESPACE \
  $ISTIO_INSTALLER_PATH/istio-control/istio-discovery/ \
  -f $ISTIO_INSTALLER_PATH/global.yaml \
  -f ${TEMP_DIR}/values-${SCENARIO}.yaml \
  | $KUBECTL2 apply -f -

# shellcheck disable=SC2086
helm template \
  --namespace $NAMESPACE \
  -n autoinject-$NAMESPACE \
  $ISTIO_INSTALLER_PATH/istio-control/istio-autoinject/ \
  -f $ISTIO_INSTALLER_PATH/global.yaml \
  -f ${TEMP_DIR}/values-${SCENARIO}.yaml \
  | $KUBECTL2 apply -f -

# shellcheck disable=SC2086
$KUBECTL2 -n $NAMESPACE rollout status deployment istio-galley
# shellcheck disable=SC2086
$KUBECTL2 -n $NAMESPACE rollout status deployment istio-pilot
# shellcheck disable=SC2086
$KUBECTL2 -n $NAMESPACE rollout status deployment istio-sidecar-injector

# shellcheck disable=SC2086
$ISTIOCTL kube-inject --context=$KUBECONTEXT2 -i $NAMESPACE -f "$BASE_DIR/$SCENARIO/setup.yaml" \
  | $KUBECTL2 -n $APPS_NAMESPACE apply -f -

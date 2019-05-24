#!/usr/bin/env bash

set -e

echo
echo " [*] Installing the 'default' scenario"
echo

SCENARIO="default"
NAMESPACE="env-$SCENARIO"
$KUBECTL2 create ns $NAMESPACE || true

$KUBECTL2 -n $NAMESPACE delete secrets -l "istio/multiCluster=true"
install_k8s_secrets "$KUBECONFIG2" "$KUBECONTEXT2" "$NAMESPACE" "$KUBECONFIG1" "$KUBECONTEXT1" "istio-system"

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

helm template \
  --namespace $NAMESPACE \
  -n galley-$NAMESPACE \
  $ISTIO_INSTALLER_PATH/istio-control/istio-config/ \
  -f $ISTIO_INSTALLER_PATH/global.yaml \
  -f ${TEMP_DIR}/values-${SCENARIO}.yaml \
  | $KUBECTL2 apply -f -

helm template \
  --namespace $NAMESPACE \
  -n pilot-$NAMESPACE \
  $ISTIO_INSTALLER_PATH/istio-control/istio-discovery/ \
  -f $ISTIO_INSTALLER_PATH/global.yaml \
  -f ${TEMP_DIR}/values-${SCENARIO}.yaml \
  | $KUBECTL2 apply -f -

helm template \
  --namespace $NAMESPACE \
  -n autoinject-$NAMESPACE \
  $ISTIO_INSTALLER_PATH/istio-control/istio-autoinject/ \
  -f $ISTIO_INSTALLER_PATH/global.yaml \
  -f ${TEMP_DIR}/values-${SCENARIO}.yaml \
  | $KUBECTL2 apply -f -

$KUBECTL2 -n $NAMESPACE rollout status deployment istio-galley
$KUBECTL2 -n $NAMESPACE rollout status deployment istio-pilot
$KUBECTL2 -n $NAMESPACE rollout status deployment istio-sidecar-injector

$ISTIOCTL kube-inject --context=$KUBECONTEXT2 -i $NAMESPACE -f "$BASE_DIR/$SCENARIO/setup.yaml" \
  | $KUBECTL2 -n $APPS_NAMESPACE apply -f -

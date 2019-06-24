#!/usr/bin/env bash

set -e

echo
echo " [*] Installing the 'locality-distribute' scenario"
echo

SCENARIO="locality-distribute"
NAMESPACE="env-$SCENARIO"
$KUBECTL2 create ns $NAMESPACE || true

$KUBECTL2 -n $NAMESPACE delete secrets -l "istio/multiCluster=true"
install_k8s_secrets "$KUBECONFIG2" "$KUBECONTEXT2" "$NAMESPACE" "$KUBECONFIG1" "$KUBECONTEXT1" "istio-system"

REGION1=$($KUBECTL1 get nodes -o jsonpath='{.items[0].metadata.labels.failure-domain\.beta\.kubernetes\.io/region}')
REGION2=$($KUBECTL2 get nodes -o jsonpath='{.items[0].metadata.labels.failure-domain\.beta\.kubernetes\.io/region}')

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

  localityLbSetting:
    distribute:
    - from: "${REGION2}/*"
      to:
        "${REGION2}/*": 75
        "${REGION1}/*": 25
pilot:
  configNamespace: ${NAMESPACE}
  env:
    PILOT_ENABLE_LOCALITY_LOAD_BALANCING: "1"
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

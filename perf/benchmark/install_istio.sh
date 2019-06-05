#!/bin/bash
set -euo pipefail
log() { echo "$1" >&2; }


log "✅ Preparing for install..."
cd istio-1.1.7
kubectl create namespace istio-system 
helm template install/kubernetes/helm/istio-init --name istio-init --namespace istio-system | kubectl apply -f -
log "Sleeping for 15 seconds while CRDs are created..."
sleep 15


log "✅ Generating installation template..."
helm template ./install/kubernetes/helm/istio --name istio --namespace istio-system \
   --set mixer.telemetry.enabled=false \
   --set mixer.policy.enabled=false \
   --set prometheus.enabled=true > istio.yaml


log "✅ Installing Istio..."
kubectl apply -f istio.yaml 
cd ..
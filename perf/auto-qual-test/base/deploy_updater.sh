#!/bin/bash
# this will look much better once we have kubectl 1.14 with kustomize support
kubectl create configmap qual-test-deployer --from-file=deploy_latest_daily.sh,../../bin/redeploy.sh --dry-run -o yaml | kubectl apply -f -
kubectl apply -f .
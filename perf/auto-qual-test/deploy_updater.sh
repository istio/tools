#!/bin/bash
# this will look much better once we have kubectl 1.14 with kustomize support
kubectl create configmap qual-test-deployer --from-file=deploy_latest_daily.sh
kubectl apply -f qual-test-update-job.yaml
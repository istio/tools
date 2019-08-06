#!/bin/bash
set -ex
WD=$(dirname $0)
WD=$(cd $WD;pwd)

deployments=$(kubectl get deployments -n service-graph00 -l app=service-graph -o jsonpath="{.items[*].metadata.name}")

for deployment in ${deployments}; do
  kubectl scale deployments -n service-graph00 ${deployment} --replicas $((1 + RANDOM % 3))
  sleep {{ .Values.configSleep }}
  # add jitter
  sleep $[ ( $RANDOM % {{ .Values.configSleep }} )  + 1 ]s
done

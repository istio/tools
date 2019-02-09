#!/bin/bash

set -xe
NUM=${NUM:?"specify the number of httpbin and sleep workloads"}

function inject_workload() {
  local deployfile="${1:?"please specify the workload deployment file"}"
  # This test uses perf/istio/values-istio-sds-auth.yaml, in which
  # Istio auto sidecar injector is not enabled.
  istioctl kube-inject -f "${deployfile}" -o temp-workload-injected.yaml
  kubectl apply -f temp-workload-injected.yaml
}

sed -e "s/httpbin-num-of-replicas/${NUM}/g" -e "s/sleep-num-of-replicas/${NUM}/g" ./httpbin_template.yaml > httpbin_sleep_deploy.yaml
inject_workload httpbin_sleep_deploy.yaml

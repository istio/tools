#!/bin/bash

set -xe
WORKLOAD_FILE=${WORKLOAD_FILE:?"specify the workload deployment file"}

function inject_workload() {
  local deployfile="${1:?"please specify the workload deployment file"}"
  # This test uses perf/istio/values-istio-sds-auth.yaml, in which
  # Istio auto sidecar injector is not enabled.
  istioctl kube-inject -f "${deployfile}" -o temp-workload-injected.yaml
  kubectl apply -f temp-workload-injected.yaml
}

inject_workload "${WORKLOAD_FILE}"

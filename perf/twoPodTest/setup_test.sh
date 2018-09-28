#!/bin/bash
set -e
NAMESPACE=${NAMESPACE:?"namespace"}

# Get pod ip range, there must be a better way, but this works.
function ip_range() {
    kubectl get pods --namespace kube-system -o wide | grep kube-dns | awk '{print $6}'|head -1 | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'
}

function run_test() {
  helm -n ${NAMESPACE} template \
	  --set excludeOutboundIPRanges=$(ip_range) \
          . > ${TMPDIR}/twopods.yaml
  echo "Wrote ${TMPDIR}/twopods.yaml"

#         kubectl apply -n ${NAMESPACE} -f ${TMPDIR}/twopods.yaml
}

run_test

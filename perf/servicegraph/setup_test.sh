#!/bin/bash
set -e
NAMESPACE=${NAMESPACE:?"namespace"}

# Get pod ip range, there must be a better way, but this works.
function ip_range() {
    kubectl get pods --namespace kube-system -o wide | grep kube-dns | awk '{print $6}'|head -1 | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'
}

function run_test() {
  YAML=$(mktemp).yml
  helm -n ${NAMESPACE} template \
	  --set excludeOutboundIPRanges=$(ip_range) \
          . > "${YAML}"
  echo "Wrote ${YAML}"

  # remove stdio rules
  kubectl --namespace istio-system delete rules stdio stdiotcp
#         kubectl apply -n ${NAMESPACE} -f ${TMPDIR}/twopods.yaml
}

run_test

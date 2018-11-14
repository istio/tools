#!/bin/bash
WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -ex
NAMESPACE=${NAMESPACE:?"namespace"}
DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org or local"}
TMPDIR=${TMPDIR:-${WD}/tmp}

mkdir -p "${TMPDIR}"

# Get pod ip range, there must be a better way, but this works.
function ip_range() {
    kubectl get pods --namespace kube-system -o wide | grep kube-dns | awk '{print $6}'|head -1 | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'
}

function run_test() {
  helm -n ${NAMESPACE} template \
	  --set excludeOutboundIPRanges=$(ip_range) \
    --set domain="${DNS_DOMAIN}" \
          . > ${TMPDIR}/twopods.yaml
  echo "Wrote ${TMPDIR}/twopods.yaml"

  # remove stdio rules
  kubectl --namespace istio-system delete rules stdio stdiotcp || true
  kubectl apply -n ${NAMESPACE} -f ${TMPDIR}/twopods.yaml
}

run_test

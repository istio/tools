#!/bin/bash
WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -x
NAMESPACE="twopods"
DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org or local"}
TMPDIR=${TMPDIR:-${WD}/tmp}
RBAC_ENABLED="false"
LINKERD_INJECT="${LINKERD_INJECT:-'disabled'}"
echo "linkerd inject is ${LINKERD_INJECT}"

mkdir -p "${TMPDIR}"

# Get pod ip range, there must be a better way, but this works.
function pod_ip_range() {
    kubectl get pods --namespace kube-system -o wide | grep kube-dns | awk '{print $6}'|head -1 | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'
}

function svc_ip_range() {
    kubectl -n kube-system get svc kube-dns --no-headers | awk '{print $3}' | awk -F '.' '{printf "%s.%s.0.0/16\n", $1, $2}'
}

function run_test() {
  helm -n ${NAMESPACE} template \
      --set rbac.enabled="${RBAC_ENABLED}" \
      --set includeOutboundIPRanges=$(svc_ip_range) \
      --set injectL="${LINKERD_INJECT}" \
      --set domain="${DNS_DOMAIN}" \
          . > ${TMPDIR}/twopods.yaml
  echo "Wrote ${TMPDIR}/twopods.yaml"

  # remove stdio rules
  kubectl apply -n ${NAMESPACE} -f ${TMPDIR}/twopods.yaml
  echo ${TMPDIR}/twopods.yaml
}

for ((i=1; i<=$#; i++)); do
    case ${!i} in
        -r|--rbac) ((i++)); RBAC_ENABLED="true"
        continue
        ;;
    esac
done
kubectl create ns ${NAMESPACE} || true
kubectl label namespace ${NAMESPACE} istio-injection=enabled --overwrite || true
run_test

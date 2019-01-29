#/bin/bash
set -ex

DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org"}

GATEWAY_URL=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

function install_all_config() {
  local DIRNAME="${1:?"output dir"}"
  local domain=${DNS_DOMAIN:-qualistio.org}
  local OUTFILE="${DIRNAME}/all_config.yaml"

  kubectl create ns test || true

  kubectl label namespace test istio-injection=enabled || true

  helm -n test template \
    --set fortioImage=fortio/fortio:latest \
    --set ingress="${GATEWAY_URL}" \
    --set domain="${domain}" . > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl -n test apply -f "${OUTFILE}"
  fi
}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

install_all_config "${WD}/tmp" $*

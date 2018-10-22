#/bin/bash
set -ex

function download() {
  local DIRNAME="$1"
  local release="$2"

  local url="https://storage.googleapis.com/istio-prerelease/daily-build/${release}/gcr.io/istio-${release}-linux.tar.gz"
  local outfile="${DIRNAME}/istio-${release}.tgz"
    
  if [[ ! -f "${outfile}" ]]; then
    wget -O "${outfile}" "${url}"
  fi

  echo "${outfile}" 
}

function install_istio() {
  local DIRNAME="${1:?"output dir"}"
  local release="${2:?"release"}"
  
  local outfile="$(download ${DIRNAME} ${release})"

  if [[ ! -d "${DIRNAME}/${release}" ]];then
      tar -xzf "${outfile}" -C "${DIRNAME}"
      mv "${DIRNAME}/istio-${release}/install/kubernetes/helm/istio" "${DIRNAME}/${release}"
      rm -Rf "${DIRNAME}/istio-${release}"
  fi

  kubectl create ns istio-system || true

  kubectl apply -f "${DIRNAME}/${release}/templates/crds.yaml"

  local FILENAME="${DIRNAME}/${release}.yml"
  helm template --name istio --namespace istio-system \
    --set global.tag=${release} \
    --set global.hub=gcr.io/istio-release \
    --values values-istio-test.yaml \
    "${DIRNAME}/${release}" > "${FILENAME}"

  if [[ -z "${DRY_RUN}" ]];then
    kubectl apply -f "${FILENAME}"
  fi
  
  echo "Wrote file ${FILENAME}"
}

function install_gateways() {
  helm template base | kubectl -n istio-system apply -f -
}

function install_all_config() {
  local DIRNAME="${1:?"output dir"}"
  local domain=${DNS_DOMAIN:-qualistio.org}
  local OUTFILE="${DIRNAME}/all_config.yaml"
  
  kubectl create ns test || true

  kubectl label namespace test istio-injection=enabled || true
  helm -n test template \
    --set fortioImage=fortio/fortio:latest \
    --set domain="v103.${domain}" allconfig > "${OUTFILE}"

  kubectl -n test apply -f "${OUTFILE}"
}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

release="${1:?"release"}"
shift

install_istio "${WD}/tmp" "${release}" $*
install_gateways

install_all_config "${WD}/tmp"

# Run this after adding a new name for ingress testing
function AddDNS() {
    local N=$1

    # Create DNS records
    # ingress103.qualistio.org.    A    300     35.239.63.185
    # *.v103.qualistio.org.    CNAME    300    ingress103.qualistio.org.

    # TODO

    gcloud dns --project=$DNS_PROJECT record-sets transaction start --zone=$DNS_ZONE

    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=${N}.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE

    gcloud dns --project=$DNS_PROJECT record-sets transaction execute --zone=$DNS_ZONE
}

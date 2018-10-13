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

  kubectl apply -f "${FILENAME}"
  
  echo "Wrote file ${FILENAME}"
}

function install_gateways() {
  helm template base | kubectl -n istio-system apply -f -
}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

release="${1:?"release"}"
shift

install_istio "${WD}/tmp" "${release}" $*
install_gateways




function testIstioSystem() {
   local FILENAME=$(mktemp).yml
   pushd $TOP/src/istio.io/istio
   helm template --name istio --namespace istio-system \
    --set global.tag=$TAG \
    --set global.hub=$HUB \
    --values tests/helm/values-istio-test.yaml \
    install/kubernetes/helm/istio  > "${FILENAME}"
    
    kubectl apply -f "${FILENAME}"

    echo "Wrote file ${FILENAME}"
   popd

}

# Install istio
function testInstall() {
    make istio-demo.yaml
    kubectl create ns istio-system
    testIstioSystem

    kubectl create ns test
    kubectl label namespace test istio-injection=enabled

    kubectl -n test apply -f samples/httpbin/httpbin.yaml
    kubectl create ns bookinfo
    kubectl label namespace bookinfo istio-injection=enabled
    kubectl -n bookinfo apply -f samples/bookinfo/kube/bookinfo.yaml
}

# Apply the helm template
function testApply() {
   local F=${1:-"istio/fortio:latest"}
   local domain=${DNS_DOMAIN:-istio.webinf.info}
   pushd $TOP/src/istio.io/istio
   helm -n test template \
    --set fortioImage=$F \
    --set domain="v10.${domain}" \
    tests/helm |kubectl -n test apply -f -
   popd
}

function testApply1() {
    testApply istio/fortio:1.0.1
}

# Setup DNS entries - currently using gcloud
# Requires DNS_PROJECT, DNS_DOMAIN and DNS_ZONE to be set
# For example, DNS_DOMAIN can be istio.example.com and DNS_ZONE istiozone.
# You need to either buy a domain from google or set the DNS to point to gcp.
# Similar scripts can setup DNS using a different provider
function testCreateDNS() {

    gcloud dns --project=$DNS_PROJECT record-sets transaction start --zone=$DNS_ZONE

  #  gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=grafana.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=prom.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=fortio2.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=pilot.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=fortio.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=fortioraw.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=bookinfo.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=httpbin.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=citadel.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE
    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=mixer.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE

    gcloud dns --project=$DNS_PROJECT record-sets transaction execute --zone=$DNS_ZONE
}

# Run this after adding a new name for ingress testing
function testAddDNS() {
    local N=$1

    gcloud dns --project=$DNS_PROJECT record-sets transaction start --zone=$DNS_ZONE

    gcloud dns --project=$DNS_PROJECT record-sets transaction add ingress10.${DNS_DOMAIN}. --name=${N}.v10.${DNS_DOMAIN}. --ttl=300 --type=CNAME --zone=$DNS_ZONE

    gcloud dns --project=$DNS_PROJECT record-sets transaction execute --zone=$DNS_ZONE
}

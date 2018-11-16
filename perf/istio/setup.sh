#/bin/bash
set -ex

DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org"}

function download() {
  local DIRNAME="$1"
  local release="$2"

  local url="https://storage.googleapis.com/istio-prerelease/daily-build/${release}/istio-${release}-linux.tar.gz"
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
      mv "${DIRNAME}/istio-${release}/install/kubernetes/helm" "${DIRNAME}/${release}"
      rm -Rf "${DIRNAME}/istio-${release}"
      helm dep update "${DIRNAME}/${release}/istio" || true
  fi

  kubectl create ns istio-system || true

  if [[ -z "${DRY_RUN}" ]];then
      kubectl apply -f "${DIRNAME}/${release}/istio/templates/crds.yaml"
  fi

  local FILENAME="${DIRNAME}/${release}.yml"

  opts="--set global.tag=${release}"
  opts+=" --set global.hub=gcr.io/istio-release"

  if [[ "${MCP}" = "1" ]];then
      opts+=" --set global.useMCP=true"
  fi

  helm template --name istio --namespace istio-system \
       ${opts} \
       --values values-istio-test.yaml \
       "${DIRNAME}/${release}/istio" > "${FILENAME}"

  # update prometheus scape interval
  sed -i 's/scrape_interval: .*$/scrape_interval: 30s/' "${FILENAME}"

  if [[ -z "${DRY_RUN}" ]];then
      kubectl apply -f "${FILENAME}"

      # remove stdio rules
      kubectl --namespace istio-system delete rules stdio stdiotcp || true
  fi

  echo "Wrote file ${FILENAME}"
}

function install_gateways() {
  local domain=${DNS_DOMAIN:-qualistio.org}
  if [[ -z "${DRY_RUN}" ]]; then
      helm template --set domain="${domain}" base | kubectl -n istio-system apply -f -
  fi
}

function install_all_config() {
  local DIRNAME="${1:?"output dir"}"
  local domain=${DNS_DOMAIN:-qualistio.org}
  local OUTFILE="${DIRNAME}/all_config.yaml"

  kubectl create ns test || true

  kubectl label namespace test istio-injection=enabled || true

  helm -n test template \
    --set fortioImage=fortio/fortio:latest \
    --set domain="${domain}" allconfig > "${OUTFILE}"

  if [[ -z "${DRY_RUN}" ]]; then
      kubectl -n test apply -f "${OUTFILE}"
  fi
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

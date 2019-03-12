#!/bin/bash
set -ex

DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org"}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

release="${1:?"release"}"
shift

function download() {
  local DIRNAME="$1"
  local release="$2"

  local url="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${release}/istio-${release}-linux.tar.gz"
  if [[ ! -z "${RELEASE_URL}" ]];then
    url="${RELEASE_URL}"
  fi
  local outfile="${DIRNAME}/istio-${release}.tgz"

  if [[ ! -f "${outfile}" ]]; then
    wget -O "${outfile}" "${url}"
  fi

  echo "${outfile}"
}

function setup_admin_binding() {
  kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account) || true
}

function install_prometheus() {
  local DIRNAME="$1" # should be like tools/perf/istio/tmp
  local PROMETHEUS_INSTALL="${DIRNAME}/../../prometheus-install"

  # Create GCP SSD Storage class for Prometheus to use. May not work outside GKE
  kubectl apply -f "${PROMETHEUS_INSTALL}/ssd-storage-class.yaml"

  helm fetch stable/prometheus-operator --untar --untardir "${PROMETHEUS_INSTALL}"

  # Store original context namespace so it can be reset at the end
  local ORIG_CTX=$(kubectl config current-context)
  local ORIG_NS=$(kubectl config get-contexts ${ORIG_CTX} --no-headers | tr -s ' ' | cut -d ' ' -f 5)
  # Prometheus operator chart doesn't respect --namespace, all objects are
  # deployed to the default namespace.
  kubectl config set-context $(kubectl config current-context) --namespace=istio-system
  helm template "${PROMETHEUS_INSTALL}/prometheus-operator/"\
    -f "${PROMETHEUS_INSTALL}/prometheus-operator-values.yaml"\
    --set-file .Values.prometheus.prometheusSpec.additionalScrapeConfigs="${PROMETHEUS_INSTALL}/prometheus-scrape-configs.yaml"\
    | kubectl apply -f -

  # Reset to original context namespace
  kubectl config set-context ${ORIG_CTX} --namespace=${ORIG_NS}
}

function install_istio() {
  local DIRNAME="${1:?"output dir"}"
  local release="${2:?"release"}"

  local outfile="$(download ${DIRNAME} ${release})"

  if [[ ! -d "${DIRNAME}/${release}" ]];then
      DN=$(mktemp -d)
      tar -xzf "${outfile}" -C "${DN}" --strip-components 1
      mv "${DN}/install/kubernetes/helm" "${DIRNAME}/${release}"
      rm -Rf ${DN}
      helm init -c
      if [[ ! ${release} =~ release-1.0-* ]];then
        local helmrepo="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${release}/charts"
        if [[ ! -z "${HELMREPO_URL}" ]];then
          helmrepo="${HELMREPO_URL}"
        fi
        helm repo add istio.io "${helmrepo}"
      fi
      helm dep update "${DIRNAME}/${release}/istio" || true
  fi

  kubectl create ns istio-system || true

  if [[ -z "${DRY_RUN}" ]];then
      # apply CRD files for istio kinds
      if [[ -f "${DIRNAME}/${release}/istio/templates/crds.yaml" ]];then
         kubectl apply -f "${DIRNAME}/${release}/istio/templates/crds.yaml"
      else
         kubectl apply -f "${DIRNAME}/${release}/istio-init/files/"
      fi
  fi

  local FILENAME="${DIRNAME}/${release}.yml"

  # if release_url is not overridden then daily builds require
  # tag and hub overrides
  if [[ -z "${RELEASE_URL}" ]];then
    opts="--set global.tag=${release}"
    opts+=" --set global.hub=gcr.io/istio-release"
  fi

  if [[ "${MCP}" != "0" ]];then
      opts+=" --set global.useMCP=true"
  fi


  local values=${VALUES:-values.yaml}

  helm template --name istio --namespace istio-system \
       ${opts} \
       --values ${values} \
       "${DIRNAME}/${release}/istio" > "${FILENAME}"

  # update prometheus scape interval
  sed -i 's/scrape_interval: .*$/scrape_interval: 30s/' "${FILENAME}"

  if [[ -z "${DRY_RUN}" ]];then
      kubectl apply -f "${FILENAME}"

      # remove stdio rules
      kubectl --namespace istio-system delete rules stdio stdiotcp || true

      install_prometheus ${DIRNAME}

  fi

  echo "Wrote file ${FILENAME}"
}

function install_gateways() {
  local domain=${DNS_DOMAIN:-qualistio.org}
  if [[ -z "${DRY_RUN}" ]]; then
      helm template --set domain="${domain}" base | kubectl -n istio-system apply -f -
  fi
}

setup_admin_binding
install_istio "${WD}/tmp" "${release}" $*
install_gateways


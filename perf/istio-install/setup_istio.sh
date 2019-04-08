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

  kubectl create ns istio-prometheus || true
  PROM_OP=tmp/prom-operator.yaml
  helm template "${PROMETHEUS_INSTALL}/prometheus-operator/" \
    --namespace istio-prometheus \
    --name prometheus \
    -f "${PROMETHEUS_INSTALL}/values.yaml" > "${PROM_OP}"

  echo "${PROM_OP}"

  kubectl apply --namespace istio-prometheus -f "${PROM_OP}"

  # delete grafana pod so it redeploys with new config
  kubectl delete pod -l app=grafana
}

function install_istio() {
  local DIRNAME="${1:?"output dir"}"
  local release="${2:?"release"}"

  local outfile="$(download ${DIRNAME} ${release})"

  if [[ ! -d "${DIRNAME}/${release}" ]]; then
      DN=$(mktemp -d)
      tar -xzf "${outfile}" -C "${DN}" --strip-components 1
      mv "${DN}/install/kubernetes/helm" "${DIRNAME}/${release}"
      tar -xzf "${DIRNAME}/${release}"/charts/istio-cni-*.tgz -C "${DIRNAME}/${release}"
      rm -Rf ${DN}
  fi

  kubectl create ns istio-system || true

  if [[ -z "${DRY_RUN}" ]]; then
    case "${CNI}" in
      "gke"|"GKE") # GKE needs a specific flag set
        helm template "${DIRNAME}/${release}/istio-cni" \
        --name=istio-cni \
        --namespace=istio-system \
        --set cniBinDir=/home/kubernetes/bin | kubectl apply -f -
       ;;
      "") ;;
      *)
        helm template "${DIRNAME}/${release}/istio-cni" \
        --name=istio-cni \
        --namespace=istio-system \
        | kubectl apply -f -
       ;;
    esac
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
    opts+=" --set global.tag=${release}"
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

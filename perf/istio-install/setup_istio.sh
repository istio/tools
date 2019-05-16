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
    curl -JLo "${outfile}" "${url}"
  fi

  echo "${outfile}"
}

function setup_admin_binding() {
  kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account) || true
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

  if [[ -z "${DRY_RUN}" ]];then
      kubectl apply -f "${FILENAME}"

      "$WD/setup_prometheus.sh" ${DIRNAME}
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

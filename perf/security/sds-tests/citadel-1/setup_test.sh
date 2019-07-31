#!/bin/bash

NAMESPACE=${NAMESPACE:?"specify the namespace for running the test"}
NUM=${NUM:?"specify the number of httpbin and sleep workloads"}
RELEASE=${RELEASE:?"specify the Istio release, e.g., release-1.1-20190208-09-16"}
RELEASETYPE=${RELEASETYPE:?"specify the Istio release type, daily, release, pre-release"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

# Download the istioctl
WD=$(dirname $0)/tmp
if [[ ! -d "${WD}" ]]; then
  mkdir $WD
fi

URL=""
case "${OSTYPE}" in
  darwin*)
    if [[ "$RELEASETYPE" == "daily" ]]; then
        URL="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${RELEASE}/istio-${RELEASE}-osx.tar.gz"
    elif [[ "$RELEASETYPE" == "release" ]]; then
        URL="https://github.com/istio/istio/releases/download/${RELEASE}/istio-${RELEASE}-osx.tar.gz"
    elif [[ "$RELEASETYPE" == "pre-release" ]]; then
        URL="https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${RELEASE}/istio-${RELEASE}-osx.tar.gz"
    else
        echo "Please specify RELEASETYPE"
    fi ;;
  linux*)
    if [[ "$RELEASETYPE" == "daily" ]]; then
        URL="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${RELEASE}/istio-${RELEASE}-linux.tar.gz"
    elif [[ "$RELEASETYPE" == "release" ]]; then
        URL="https://github.com/istio/istio/releases/download/${RELEASE}/istio-${RELEASE}-linux.tar.gz"
    elif [[ "$RELEASETYPE" == "pre-release" ]]; then
        URL="https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${RELEASE}/istio-${RELEASE}-linux.tar.gz"
    else
        echo "Please specify RELEASETYPE"
    fi ;;
  *) echo "unsupported: ${OSTYPE}" ;;
esac

curl -JLo "$WD/istio-${RELEASE}.tar.gz" "${URL}"
tar xfz ${WD}/istio-${RELEASE}.tar.gz -C $WD

function inject_workload() {
  local deployfile="${1:?"please specify the workload deployment file"}"
  # This test uses perf/istio/values-istio-sds-auth.yaml, in which
  # Istio auto sidecar injector is not enabled.
  $WD/istio-${RELEASE}/bin/istioctl kube-inject -f "${deployfile}" -o temp-workload-injected.yaml
  kubectl apply -n ${NAMESPACE} -f temp-workload-injected.yaml --cluster ${CLUSTER}
}

TEMP_DEPLOY_NAME="temp_httpbin_sleep_deploy.yaml"
helm template --set replicas="${NUM}" ../../workload-deployments/ > "${TEMP_DEPLOY_NAME}"

kubectl create ns ${NAMESPACE} --cluster ${CLUSTER}

inject_workload ${TEMP_DEPLOY_NAME}

echo "Wait 60 seconds for the deployment to be ready ..."
sleep 60


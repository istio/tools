#!/bin/bash

set +x
set -e

echo "Building standalone binary"

export NAMESPACE=${NAMESPACE:-'twopods-istio'}

if [ ! -d dist/ ]
then
    rm requirements.txt || true
    python3 -m venv  procstatenv
    source procstatenv/bin/activate
    pip3 install prometheus-client psutil
    # We strip a line because of a bug in pip freeze
    pip freeze | grep -v "pkg-resources" > requirements.txt
    # We build on a docker to make sure we produce a compatible binary
    # (we need to make sure to build it with a compatible glibc version)
    # TODO(oschaaf): is it OK to use this docker image?
    docker run -v "$(pwd):/src/" cdrx/pyinstaller-linux:python3
fi

echo "Deploying standalone binary"

kubectl get pods --namespace twopods-istio --no-headers --field-selector=status.phase=Running -o name | while read pod
do
    # Strip the pod/ prefix we get for free
    pod=${pod#"pod/"}
    echo "Installing to ${pod}"
    kubectl --namespace ${NAMESPACE} exec ${pod} -c istio-proxy -- rm -rf /etc/istio/proxy/procstat
    kubectl --namespace ${NAMESPACE} cp ./ ${pod}:/etc/istio/proxy/procstat -c istio-proxy
    echo "Fire service in ${pod}"
    # Stop the existing service instance, if any
    kubectl --namespace ${NAMESPACE} exec ${pod} -c istio-proxy -- pkill -f prom || true
    # Fix, this neesd the kubectl command to stay running on the machine running this script 
    kubectl --namespace ${NAMESPACE} exec ${pod} -c istio-proxy /etc/istio/proxy/procstat/dist/linux/prom/prom &
done

echo "proc stat sampling deployed"



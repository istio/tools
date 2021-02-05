#!/bin/bash

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -eux

red='\[\033[0;31m\]'
clr='\[\033[0m\]'

VM_APP="${VM_APP:?}"
VM_NAME="${VM_NAME:-${VM_APP}}"
VM_NAMESPACE="${VM_NAMESPACE:?}"
VERSION="${VERSION:?"version, like 1.10-alpha.45c5661eb8c96cebe8fcb467b4c1be3262b0de4c"}"
PROJECT="${PROJECT:-mixologist-142215}"
ZONE="${ZONE:-us-central1-c}"
WORK_DIR=/tmp/vm
SERVICE_ACCOUNT=default
export CLOUDSDK_COMPUTE_ZONE="${ZONE}"
export CLOUDSDK_CORE_PROJECT="${PROJECT}"

mkdir -p "${WORK_DIR}"

docker-copy() {
    image="${1:?image}"
    src="${2:?src}"
    dst="${3:?dst}"
    docker create --rm --name temp-docker-copy "${image}"
    docker cp temp-docker-copy:"${src}" "${dst}"
    docker stop temp-docker-copy
    docker rm temp-docker-copy
}

gcloud compute instances describe "${VM_APP:?}" > /dev/null 2>&1 && { echo "${red}Instance already configured! Warning: script will not update VM.${clr}"; exit 0; }

gcloud compute instances create "${VM_NAME}" \
  --image-family debian-10 --image-project debian-cloud \
  --machine-type e2-standard-2

kubectl create namespace "${VM_NAMESPACE}" || true
kubectl create serviceaccount "${SERVICE_ACCOUNT}" -n "${VM_NAMESPACE}" || true

kubectl get cm -n "${VM_NAMESPACE}" service-graph-config -ojsonpath='{.data.service-graph}' > "${WORK_DIR}"/service-graph.yaml

istioctl x workload group create --name "${VM_APP}" --namespace "${VM_NAMESPACE}" --labels app="${VM_APP}" --serviceAccount "${SERVICE_ACCOUNT}" >  "${WORK_DIR}"/workloadgroup.yaml
cat <<EOF > "${WORK_DIR}"/workloadgroup.yaml
apiVersion: networking.istio.io/v1alpha3
kind: WorkloadGroup
metadata:
  name: "${VM_APP}"
  namespace: "${VM_NAMESPACE}"
spec:
  metadata:
    labels:
      app: "${VM_APP}"
  template:
    serviceAccount: "${SERVICE_ACCOUNT}"
  probe:
    httpGet:
      path: /metrics
      port: 8080
    initialDelaySeconds: 5
    periodSeconds: 5
EOF

kubectl --namespace "${VM_NAMESPACE}" apply -f "${WORK_DIR}"/workloadgroup.yaml

istioctl x workload entry configure -f "${WORK_DIR}"/workloadgroup.yaml -o "${WORK_DIR}" --autoregister

# Wait until we can ssh
sleep 15

cat <<EOF > "${WORK_DIR}"/isotope.service
[Unit]
Description=Isotope
After=network.target
StartLimitIntervalSec=0
[Service]
Type=simple
Restart=always
Environment="SERVICE_NAME=${VM_APP}"
RestartSec=1
ExecStart=/usr/bin/isotope_service --max-idle-connections-per-host=32

[Install]
WantedBy=multi-user.target
EOF

docker-copy gcr.io/istio-testing/isotope:0.0.1 /usr/local/bin/isotope_service  "${WORK_DIR}"/isotope_service

gcloud compute scp "${WORK_DIR}"/* "${VM_APP}":
gcloud compute ssh  "${VM_APP}" -- sudo bash -c "\"
mkdir -p /etc/certs /var/run/secrets/tokens /etc/istio/config/ /etc/istio/proxy /etc/config
curl -LO https://storage.googleapis.com/istio-build/dev/${VERSION}/deb/istio-sidecar.deb
sudo dpkg -i istio-sidecar.deb
cp root-cert.pem /etc/certs/root-cert.pem
cp istio-token /var/run/secrets/tokens/istio-token
cp cluster.env /var/lib/istio/envoy/cluster.env
cp mesh.yaml /etc/istio/config/mesh
cp service-graph.yaml /etc/config/service-graph.yaml
cp isotope_service /usr/bin/isotope_service
cp isotope.service /etc/systemd/system/isotope.service
chmod 777 /etc/config/service-graph.yaml
cat hosts >> /etc/hosts
chown -R istio-proxy /var/lib/istio /etc/certs /etc/istio/proxy /etc/istio/config /var/run/secrets /etc/certs/root-cert.pem
systemctl start istio
systemctl start isotope
\""

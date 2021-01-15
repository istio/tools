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

export proj="${proj:-istio-testing}"
export zone="${zone:-us-central1-a}"
export cluster1="${cluster1:-cluster1}"
export cluster2="${cluster2:-cluster2}"
export RELEASE="${RELEASE:-release-1.1-20190209-09-16}"

# We must create clusters sequentially without specifying --async, otherwise will fail.
function create_clusters() {
  # shellcheck disable=SC2140
  scope="https://www.googleapis.com/auth/compute","https://www.googleapis.com/auth/devstorage.read_only",\
"https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring",\
"https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly",\
"https://www.googleapis.com/auth/trace.append"
  # shellcheck disable=SC2086
	gcloud container clusters create $cluster1 --zone $zone --username "admin" \
--machine-type "n1-standard-2" --image-type "COS" --disk-size "100" \
--scopes $scope --num-nodes "4" --network "default" --enable-cloud-logging --enable-cloud-monitoring --enable-ip-alias
  # shellcheck disable=SC2086
	gcloud container clusters create $cluster2 --zone $zone --username "admin" \
--machine-type "n1-standard-2" --image-type "COS" --disk-size "100" \
--scopes $scope \
--num-nodes "4" --network "default" --enable-cloud-logging --enable-cloud-monitoring --enable-ip-alias
}

function setup_admin_binding() {
  # shellcheck disable=SC2046
  kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account) || true
}

function create_cluster_admin() {
	# shellcheck disable=SC2086
	gcloud container clusters get-credentials $cluster1 --zone $zone
	# shellcheck disable=SC2086
	gcloud container clusters get-credentials $cluster2 --zone $zone
	kubectl config use-context "gke_${proj}_${zone}_${cluster1}"
	setup_admin_binding
	kubectl config use-context "gke_${proj}_${zone}_${cluster2}"
	setup_admin_binding
}

# this will create the firewalls to allow traffic from all clusters in the project to all project
# might be overkill given we only need two clusters working...
function create_firewall() {
	function join_by { local IFS="$1"; shift; echo "$*"; }
	ALL_CLUSTER_CIDRS=$(gcloud container clusters list --format='value(clusterIpv4Cidr)' | sort | uniq)
	# shellcheck disable=SC2046
	# shellcheck disable=SC2116
	ALL_CLUSTER_CIDRS=$(join_by , $(echo "${ALL_CLUSTER_CIDRS}"))
	ALL_CLUSTER_NETTAGS=$(gcloud compute instances list --format='value(tags.items.[0])' | sort | uniq)
	# shellcheck disable=SC2046
	# shellcheck disable=SC2116
	ALL_CLUSTER_NETTAGS=$(join_by , $(echo "${ALL_CLUSTER_NETTAGS}"))
	gcloud compute firewall-rules create istio-multicluster-test-pods \
		--allow=tcp,udp,icmp,esp,ah,sctp \
		--direction=INGRESS \
		--priority=900 \
		--source-ranges="${ALL_CLUSTER_CIDRS}" \
		--target-tags="${ALL_CLUSTER_NETTAGS}" --quiet
}

# TODO: reuse from perf/istio/setup.sh
function download() {
  local DIRNAME="$1"
	# shellcheck disable=SC2086
	rm -rf $DIRNAME && mkdir $DIRNAME
	https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/
  local url="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${RELEASE}/istio-${RELEASE}-linux.tar.gz"
	# local url="https://storage.googleapis.com/istio-prerelease/daily-build/release-1.1-20190212-09-16/istio-release-1.1-20190212-09-16-linux.tar.gz"
  local outfile="${DIRNAME}/istio-${RELEASE}.tgz"

  if [[ ! -f "${outfile}" ]]; then
    wget –quiet -O "${outfile}" "${url}"
  fi

  echo "${outfile}"
}

function install_istio() {
	kubectl config use-context "gke_${proj}_${zone}_${cluster1}"
	# shellcheck disable=SC2086
	istio_tar=$(download ./tmp $RELEASE)
	# shellcheck disable=SC2086
	tar xf $istio_tar -C ./tmp
	# shellcheck disable=SC2164
	# shellcheck disable=SC2086
	pushd tmp/istio-${RELEASE}
	# shellcheck disable=SC2086
	for i in install/kubernetes/helm/istio-init/files/crd*yaml; do kubectl apply -f $i; done
	helm repo add istio.io "https://storage.googleapis.com/istio-prerelease/daily-build/release-1.9/charts"
	helm dep update install/kubernetes/helm/istio
	helm template install/kubernetes/helm/istio --name istio --namespace istio-system > ./istio_master.yaml
	kubectl create ns istio-system
	kubectl apply -f ./istio_master.yaml
	kubectl label namespace default istio-injection=enabled
	# shellcheck disable=SC2164
	popd
}

# TODO: ensure the command succeed with expected output
function install_istio_remote() {
kubectl config use-context "gke_${proj}_${zone}_${cluster1}"
# shellcheck disable=SC2164
# shellcheck disable=SC2086
pushd tmp/istio-${RELEASE}
# Wait for remote to be ready
# shellcheck disable=SC2155
export PILOT_POD_IP=$(kubectl -n istio-system get pod -l istio=pilot -o jsonpath='{.items[0].status.podIP}')
# shellcheck disable=SC2155
export POLICY_POD_IP=$(kubectl -n istio-system get pod -l istio-mixer-type=policy -o jsonpath='{.items[0].status.podIP}')
# shellcheck disable=SC2155
export TELEMETRY_POD_IP=$(kubectl -n istio-system get pod -l istio-mixer-type=telemetry -o jsonpath='{.items[0].status.podIP}')
# shellcheck disable=SC2155
export ZIPKIN_POD_IP=$(kubectl -n istio-system get pod -l app=jaeger -o jsonpath='{range .items[*]}{.status.podIP}{end}')

# shellcheck disable=SC2086
helm template install/kubernetes/helm/istio --namespace istio-system \
--name istio-remote \
--values install/kubernetes/helm/istio/values-istio-remote.yaml \
--set global.remotePilotAddress=${PILOT_POD_IP} \
--set global.remotePolicyAddress=${POLICY_POD_IP} \
--set global.remoteTelemetryAddress=${TELEMETRY_POD_IP} \
> istio-remote.yaml

  # switch to the remote cluster.
  kubectl config use-context "gke_${proj}_${zone}_${cluster2}"
	kubectl create ns istio-system
	kubectl apply -f ./istio-remote.yaml
	kubectl label namespace default istio-injection=enabled
	# shellcheck disable=SC2164
	popd
}

function register_remote_cluster() {
# shellcheck disable=SC2164
# shellcheck disable=SC2086
pushd tmp/istio-${RELEASE}
# shellcheck disable=SC2155
export WORK_DIR=$(pwd)
kubectl config use-context "gke_${proj}_${zone}_${cluster2}"
CLUSTER_NAME=$(kubectl config view --minify=true -o "jsonpath={.clusters[].name}")
# k8s secrete naming requirements.
CLUSTER_NAME=${CLUSTER_NAME//[_]/.}
export KUBECFG_FILE=${WORK_DIR}/${CLUSTER_NAME}
SERVER=$(kubectl config view --minify=true -o "jsonpath={.clusters[].cluster.server}")
NAMESPACE=istio-system
SERVICE_ACCOUNT=istio-multi
SECRET_NAME=$(kubectl get sa ${SERVICE_ACCOUNT} -n ${NAMESPACE} -o jsonpath='{.secrets[].name}')
# shellcheck disable=SC2086
CA_DATA=$(kubectl get secret ${SECRET_NAME} -n ${NAMESPACE} -o "jsonpath={.data['ca\.crt']}")
# shellcheck disable=SC2086
TOKEN=$(kubectl get secret ${SECRET_NAME} -n ${NAMESPACE} -o "jsonpath={.data['token']}" | base64 --decode)

# shellcheck disable=SC2086
cat <<EOF > ${KUBECFG_FILE}
apiVersion: v1
clusters:
   - cluster:
       certificate-authority-data: ${CA_DATA}
       server: ${SERVER}
     name: ${CLUSTER_NAME}
contexts:
   - context:
       cluster: ${CLUSTER_NAME}
       user: ${CLUSTER_NAME}
     name: ${CLUSTER_NAME}
current-context: ${CLUSTER_NAME}
kind: Config
preferences: {}
users:
   - name: ${CLUSTER_NAME}
     user:
       token: ${TOKEN}
EOF

	# now switch to the master cluster
	kubectl config use-context "gke_${proj}_${zone}_${cluster1}"
	# shellcheck disable=SC2086
	kubectl create secret generic ${CLUSTER_NAME} --from-file ${KUBECFG_FILE} -n ${NAMESPACE}
	# shellcheck disable=SC2086
	kubectl label secret ${CLUSTER_NAME} istio/multiCluster=true -n ${NAMESPACE}
	# shellcheck disable=SC2164
	popd
}

# Deploy bookinfo in two clusters.
function deploy_bookinfo() {
	# shellcheck disable=SC2164
	# shellcheck disable=SC2086
	pushd tmp/istio-${RELEASE}
	kubectl config use-context "gke_${proj}_${zone}_${cluster1}"
	kubectl apply -f samples/bookinfo/platform/kube/bookinfo.yaml
	kubectl apply -f samples/bookinfo/networking/bookinfo-gateway.yaml
	kubectl delete deployment reviews-v3
	kubectl config use-context "gke_${proj}_${zone}_${cluster2}"
	kubectl apply -f ../../reviews-v3.yaml
	# shellcheck disable=SC2164
	popd
}

function get_verify_url() {
	kubectl config use-context "gke_${proj}_${zone}_${cluster1}"
	host=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
	echo "visiting http://${host}/productpage several times to verify all versions reviews services are available."
}

# TODO: bookinfo is not mentioned in the new guide.
function cleanup() {
	gcloud compute firewall-rules delete istio-multicluster-test-pods -q
	# shellcheck disable=SC2086
	gcloud container clusters delete ${cluster1} ${cluster2} -q
}

function do_all() {
	create_clusters
	create_cluster_admin
	create_firewall
	install_istio
	# Really workaround, remote istio cluster may not be ready.
	sleep 60 && install_istio_remote
	register_remote_cluster
	deploy_bookinfo
	get_verify_url
}

if [[ $# -ne 1 ]]; then
  echo "Usage: ./setup.sh cleanup | setup"
  return
fi

case $1 in
  setup)
	   do_all
		 ;;

	cleanup)
	  cleanup
		;;
esac

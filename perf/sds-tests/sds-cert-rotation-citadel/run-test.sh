# Create a cluster for testing
# Set PROJECT as the name of your GCP project
export PROJECT="endpoints-jenkins"
# Set CLUSTER as the name of your newly created GKE cluster 
export CLUSTER="istio-cluster-test-sds-enabled"
# Set ZONE as the zone of your newly created GKE cluster 
export ZONE="us-central1-a"
# Set RELEASE_NAME as the release name you are going to test
export RELEASE_NAME=release-1.1-20190129-09-16

gcloud container clusters get-credentials ${CLUSTER} --zone ${ZONE} --project ${PROJECT} 

kubectl create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user="$(gcloud config get-value core/account)"

wget https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/$RELEASE_NAME/istio-${RELEASE_NAME}-linux.tar.gz

#TODO: may need to delete previously downloaded release, if any 
tar xfz istio-${RELEASE_NAME}-linux.tar.gz

cd istio-${RELEASE_NAME}

helm dep update --skip-refresh install/kubernetes/helm/istio
cat install/kubernetes/namespace.yaml > istio-auth.yaml
cat install/kubernetes/helm/istio-init/files/crd-* >> istio-auth.yaml
helm template \
    --name=istio \
    --namespace=istio-system \
    --set global.mtls.enabled=true \
    --set global.controlPlaneSecurityEnabled=true \
    --set nodeagent.env.SECRET_GRACE_DURATION=1m \
    --set nodeagent.env.SECRET_JOB_RUN_INTERVAL=30s \
    --set nodeagent.env.SECRET_TTL=2m \
    --values install/kubernetes/helm/istio/values-istio-sds-auth.yaml \
    install/kubernetes/helm/istio >> istio-auth.yaml

# Deploy the Istio with SDS enabled. In this test, SDS goes to Citadel.
kubectl create -f istio-auth.yaml

echo "Wait 30 seconds for Istio to be ready..."
sleep 30s


istioctl kube-inject -f samples/httpbin/httpbin.yaml > httpbin-injected.yaml
# Deploy the example backend.
kubectl apply -f httpbin-injected.yaml

echo "Let the workload run 5 minutes to have a few cert. rotations."
sleep 5m

# List Node Agent’s pods:
kubectl get pod -n istio-system -l app=nodeagent -o jsonpath={.items..metadata.name}

# View Node Agent's cert. rotation logs through "kubectl logs -n istio-system NODE-AGENT-POD-NAME"


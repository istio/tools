git clone https://github.com/istio/tools
cd tools/perf/istio-install
helm init --client-only
TARGET_VERSION=$1
LATEST_BUILD=$(curl -L https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/release-$TARGET_VERSION-latest.txt)
HELMREPO_URL=https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/$LATEST_BUILD/charts/ RELEASE_URL=https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/$LATEST_BUILD/istio-$LATEST_BUILD-linux.tar.gz DNS_DOMAIN=v11.qualistio.org ./setup_istio.sh $TARGET_VERSION

NAMESPACES=$(kubectl get namespace -l istio-injection=enabled --no-headers -o custom-columns=":metadata.name")

# trigger redeploy on services to get new sidecars
# based on the "patch deployment" strategy in this comment:
# https://github.com/kubernetes/kubernetes/issues/13488#issuecomment-372532659
# requires jq

# $1 is a valid namespace
function refresh-all-pods() {
  echo
  DEPLOYMENT_LIST=$(kubectl -n $1 get deployment --no-headers -o custom-columns=":metadata.name")
  echo "Refreshing pods in all Deployments"
  for deployment_name in $DEPLOYMENT_LIST ; do
    TERMINATION_GRACE_PERIOD_SECONDS=$(kubectl -n $1 get deployment "$deployment_name" -o json|jq .spec.template.spec.terminationGracePeriodSeconds)
    if [ "$TERMINATION_GRACE_PERIOD_SECONDS" -eq 30 ]; then
      TERMINATION_GRACE_PERIOD_SECONDS='31'
    else
      TERMINATION_GRACE_PERIOD_SECONDS='30'
    fi
    patch_string="{\"spec\":{\"template\":{\"spec\":{\"terminationGracePeriodSeconds\":$TERMINATION_GRACE_PERIOD_SECONDS}}}}"
    kubectl -n $1 patch deployment $deployment_name -p $patch_string
  done
  echo
}

for namespace in $NAMESPACES ; do
    refresh-all-pods namespace
done
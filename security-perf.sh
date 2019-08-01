#!/bin/bash

# Make sure to connect to the right cluster first.
# Usage: source security-perf.sh 1.1.12 pre-release 1

export RELEASE=$1
export RELEASETYPE=$2
export DNS_DOMAIN=perf.citadel.com
export NAMESPACE="test-ns" 
export NUM=10
export CLUSTER=$(kubectl config current-context)
TEST_TYPE=$3


case "$3" in 
"1")
	echo "SDS normal perf test"
	EXTRA_VALUES="values-istio-sds-auth.yaml"
	TEST_LOC="perf/security/sds-tests/citadel-1"
	TEST_TYPE_NAME="sds-normal"
	;;
"2")
	echo "SDS redeploy perf test"
	EXTRA_VALUES="values-istio-sds-auth.yaml"
	TEST_LOC="perf/security/sds-tests/citadel-2"
	TEST_TYPE_NAME="sds-redeploy"
	;;
"3")
	echo "File-mount normal perf test"
	EXTRA_VALUES="values-istio-non-sds.yaml"
	TEST_LOC="perf/security/file-mount-tests/non-sds-1"
	TEST_TYPE_NAME="file-mount-normal"
	;;
"4")
	echo "File-mount redeploy perf test"
	EXTRA_VALUES="values-istio-non-sds.yaml"
	TEST_LOC="perf/security/file-mount-tests/non-sds-2"
	TEST_TYPE_NAME="file-mount-redeploy"
	;;
esac 

echo "Cleaning resources..."
kubectl delete namespace istio-system
kubectl delete namespace istio-prometheus
kubectl delete namespace test-ns

echo "Deploying Istio..."
cd perf/istio-install/
source setup_istio_release.sh "$RELEASE $RELEASETYPE"
 
echo "Sleep 90 seconds for Istio to be ready"
sleep 90

echo "Running test..."
cd -
cd "$TEST_LOC"
./setup_test.sh "2>&1 | tee $RELEASETYPE-$RELEASE-perf-test-$TEST_TYPE_NAME.txt"
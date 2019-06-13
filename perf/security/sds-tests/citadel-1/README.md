## Test description
This is a certificate rotation test that uses SDS with Citadel.
The certificate rotation is tested through creating a number of
httpbin and sleep workloads (the number of workloads is an input parameter of the script),
which request for certificates at a customizable interval (e.g., every 1 minute),
thereby creating the certificate rotation load. 

In this test, *controlPlaneSecurity* is disabled. To enable 
*controlPlaneSecurity*, set *controlPlaneSecurity* as *true* in
*ROOT-OF-REPO/perf/istio/values-istio-sds-auth.yaml*.

## To run the SDS test that goes to Citadel
- Create a GKE cluster and set it as the current cluster.
Here this test is ran on the cluster *sds-citadel-cert-rotation-1*
on GCP project *istio-security-testing*.
You may use `kubectl config current-context` to confirm that your newly created cluster
is set as the current cluster.

- Deploy Istio:
Let the root directory of this repo be *ROOT-OF-REPO*.
Run the following commands:
```
  cd ROOT-OF-REPO/perf/istio-install
  DNS_DOMAIN=your-example-domain VALUES=values-istio-sds-auth.yaml ./setup_istio.sh release-1.1-20190221-09-16
```  
You may replace the Istio release
in the command to the Istio release to test.
After running the above script to deploy Istio, wait a moment for Istio to be ready.

- Deploy workloads that request for certificates through SDS.
*RELEASE* variable specifies the Istio release.
*RELEASETYPE* variable specifies the type Istio release, daily, pre-release and release. 
*NAMESPACE* variable specifies the k8s namespace for testing.
*NUM* variable specifies the number of httpbin and sleep workloads.
*CLUSTER* variable specifies the cluster for running the test
(the list of clusters can be viewed through "kubectl config get-contexts").
The following example command will deploy 10 httpbin and sleep workloads in
a namespace called *test-ns*.
Note: the number of workloads can be ran depends on the size of your cluster.
```
  cd ROOT-OF-REPO/perf/security/sds-tests/citadel-1
  RELEASETYPE=daily RELEASE=release-1.1-20190221-09-16 NAMESPACE=test-ns NUM=10 CLUSTER=gke_istio-security-testing_us-central1-a_release-12-qualify-citadel-1 ./setup_test.sh
```
To test against a release or pre-release. Choose one the the following commands to set up test. 
```
  RELEASETYPE=release RELEASE=1.1.7 NAMESPACE=test-ns NUM=10 CLUSTER=gke_istio-security-testing_us-central1-a_release-12-qualify-citadel-1 ./setup_test.sh
```
or
```
  RELEASETYPE=pre-release RELEASE=1.1.7 NAMESPACE=test-ns NUM=10 CLUSTER=gke_istio-security-testing_us-central1-a_release-12-qualify-citadel-1 ./setup_test.sh
```
Wait a moment for the deployment to be ready. Then view the logs of Node Agents.
The Node Agents can be listed through
the command:
```
 kubectl get pod -n istio-system | grep nodeagent
``` 
You should see the following log entries in some of the Node Agents that show
SDS pushing certificates to the Envoy of the example workload. When a certificate
issuance or rotation occurs, the following log entries are generated.
```
  info    SDS: push root cert from node agent to proxy
  info    SDS: push key/cert pair from node agent to proxy
```

- The *setup_test.sh* script will call *collect_stats.sh* to test certificate
rotations and mTLS by curl httpbin from sleep (you may also execute *collect_stats.sh*
separately after the deployment by run `NAMESPACE=test-ns ./collect_stats.sh`). 
The following output will be displayed.
```
  Out of 1 curl, 1 succeeded.
  Out of 2 curl, 2 succeeded.
  Out of 3 curl, 3 succeeded.
  Out of 4 curl, 4 succeeded.
  Out of 5 curl, 5 succeeded.
  Out of 6 curl, 6 succeeded.
  ...
```

- After testing, you may delete the Istio and example workload created for this test
by running the following command. The namespace in the command line is the namespace
created for testing.
```
  NAMESPACE=test-ns CLUSTER=gke_istio-security-testing_us-central1-a_release-12-qualify-citadel-1 ./cleanup.sh
```

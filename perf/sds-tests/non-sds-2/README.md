## Test description
This is a certificate rotation test with SDS disabled.
The certificate rotation is tested through creating a number of
httpbin and sleep workloads (the number of workloads is an input parameter of the script),
which request for certificates at a customizable interval (e.g., every 1 minute),
thereby creating the certificate rotation load. This test also
periodically delete pods and create pods to test whether Citadel Agent
properly releases resources and handles new certificate requests.

In this test, *controlPlaneSecurity* is disabled. To enable 
*controlPlaneSecurity*, set *controlPlaneSecurity* as *true* in
*ROOT-OF-REPO/perf/istio/values-istio-non-sds.yaml*.

## To run the certificate rotation test with SDS disabled:
- Create a GKE cluster and set it as the current cluster.
Here this test is ran on the cluster *no-sds-cert-rotation-2*
on GCP project *istio-security-testing*.
You may use `kubectl config current-context` to confirm that your newly created cluster
is set as the current cluster.

- Deploy Istio:
Let the root directory of this repo be *ROOT-OF-REPO*.
Run the following commands:
```
  cd ROOT-OF-REPO/perf/istio
  DNS_DOMAIN=your-example-domain VALUES=values-istio-non-sds.yaml ./setup.sh release-1.1-20190221-09-16
```  
You may replace the Istio release
in the command to the Istio release to test.
After running the above script to deploy Istio, wait a moment for Istio to be ready.

- Deploy workloads that request for certificates. *RELEASE* variable
specifies the Istio release. *NAMESPACE* variable specifies the k8s namespace for testing.
*NUM* variable specifies the number of httpbin and sleep workloads.
The following example command will deploy 3 httpbin and sleep workloads in
a namespace called *test-ns*.
Note: the number of workloads can be ran depends on the size of your cluster.
```
  cd ROOT-OF-REPO/perf/sds-tests/non-sds-2
  RELEASE=release-1.1-20190221-09-16 NAMESPACE=test-ns NUM=10 ./setup_test.sh
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
  NAMESPACE=test-ns ./cleanup.sh
```

# Test description

This is a certificate rotation test that uses SDS with Istiod.
The certificate rotation is tested through creating a number of
httpbin and sleep workloads (the number of workloads is an input parameter of the script),
which request for certificates at a customizable interval (e.g., every 1 minute),
thereby creating the certificate rotation load.

This test creates two namespaces with equal number of workloads.
One namespace has a name prefix "static" and the other has a name prefix "dynamic".
Workloads in the latter are periodically redeployed to test SDS workflow is properly recreated.
There is a DestinationRule that requires a new connection for each request, which is to test mTLS
handshake using rotated key and cert.

## To run the SDS test that goes to Citadel

- Create a GKE cluster and set it as the current cluster.
Here this test is ran on the cluster *istio-testing*
on GCP project *istio-security-testing*.
You may use `kubectl config current-context` to confirm that your newly created cluster
is set as the current cluster.

- Deploy Istio:
Let the root directory of this repo be *ROOT-OF-REPO*.
Run the following commands:

    ```bash
    cd ROOT-OF-REPO/perf/istio-install
    DNS_DOMAIN=your-example-domain EXTRA_VALUES=values-istio-sds-auth.yaml ./setup_istio.sh release-1.5.1
    ```

    You may replace the Istio release
    in the command to the Istio release to test.
    After running the above script to deploy Istio, wait a moment for Istio to be ready.

- Deploy workloads that request for certificates through SDS.
*NUM* variable specifies the number of httpbin/sleep workloads. For example, `NUM=100` will
create 100 httpbin and 100 sleep workloads in each namespace, and that is 400 workloads in total.
*CLUSTER* variable specifies the cluster for running the test
(the list of clusters can be viewed through "kubectl config get-contexts").
The following example command will deploy two namespace, with 10 httpbin and sleep workloads in
each namespace.
Note: the number of workloads can be ran depends on the size of your cluster.

    ```bash
    NUM=100 CLUSTER=gke_istio-security-testing_us-central1-a_istio-testing ./setup_test.sh
    ```

The log in sleep container shows number of requests sent to httpbin and number of successful responses.

To download a specific version of istioctl and deploy the test worloads using that istioctl binary.

    ```bash
    NUM=100 CLUSTER=gke_istio-security-testing_us-central1-a_istio-testing ./setup_test.sh 1.5.1 pre-release
    ```

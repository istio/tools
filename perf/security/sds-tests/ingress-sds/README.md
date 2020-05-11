# Test description

This is a TLS ingress gateway SDS test.
The test sets up a number of TLS ingress gateway for a group of httpbin services.
The test creates a group of sleep pods, where each sleep sends HTTPS requests to a httpbin
service periodically. For example, sleep-1 sends HTTPS requests to httpbin-1.example.com,
and sleep-2 sends HTTPS requests to httpbin-2.example.com. The sleep and httpbin are deployed
in a test namespace with prefix "httpbin".

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
*NUM* variable specifies the number of TLS gateway and client workloads. For example, `NUM=100` will
create 100 TLS ingress gateway and 100 ingress secrets, and 100 sleep workloads as clients.
*CLUSTER* variable specifies the cluster for running the test
(the list of clusters can be viewed through "kubectl config get-contexts").
The following example command will deploy a test namespace, with 10 TLS ingress gateway and 10 sleep workloads in
the namespace.
Note: the number of workloads can be ran depends on the size of your cluster.

    ```bash
    NUM=100 CLUSTER=gke_istio-security-testing_us-central1-a_istio-testing ./setup_test.sh
    ```

The log in sleep container shows number of requests sent to httpbin and number of successful responses.

To download a specific version of istioctl and deploy the test worloads using that istioctl binary.

    ```bash
    NUM=100 CLUSTER=gke_istio-security-testing_us-central1-a_istio-testing ./setup_test.sh 1.5.1 pre-release
    ```

To delete the ingress secrets, use util script cleanup_ingress_secrets.sh. The example below deletes
100 ingress secrets. 

    ```bash
    NUM=100 CLUSTER=gke_istio-security-testing_us-central1-a_istio-testing ./cleanup_ingress_secrets.sh
    ```

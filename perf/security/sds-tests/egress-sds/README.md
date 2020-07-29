# Test description

This is a TLS egress gateway SDS test.
The test sets up a number of TLS egress gateways for a group of nginx services.
The test creates a group of sleep pods, where each sleep sends HTTP requests to an nginx
service periodically. For example, sleep-1 sends HTTP requests to my-nginx-1,
and sleep-2 sends HTTPS requests to my-nginx-2. The sleep pods are deployed in clientns namespace and nginx services are deployed
in a namespace called mesh-external".

## To run the SDS test originating mTLS at gateway

- Create a GKE cluster and set it as the current cluster.
Here this test is run on the cluster *istio-testing*
on GCP project *istio-security-testing*.
You may use `kubectl config current-context` to confirm that your newly created cluster
is set as the current cluster.

- Deploy Istio:
  istio-egressgateway must be enabled!

- Deploy workloads that request for certificates through SDS.
*NUM* variable specifies the number of client workloads and nginx services to be deployed. For example, `NUM=100` will
create 100 mutual TLS egress gateways and 100 egress secrets, and 100 sleep workloads in clients. We also configure 100 `DestinationRule` 
deployments to apply traffic policies for outgoing request to the 100 `nginx` instances running in `mesh-external` namespace.
*CLUSTER* variable specifies the cluster for running the test
(the list of clusters can be viewed through "kubectl config get-contexts").

Note: the number of workloads that can be ran depends on the size of your cluster.

    ```bash
    NUM=100 CLUSTER=gke_istio-security-testing_us-central1-a_istio-testing ./setup_test.sh
    ```

The logs for sleep containers shows number of requests that are sent to nginx and number of successful responses received.
The logs for istio-egressgateway also show the GET requests that are routed from sleep pod's in clientns namespace to corresponding
nginx pod's in mesh-external namespace

To download a specific version of istioctl and deploy the test worloads using that istioctl binary use:

    ```bash
    NUM=100 CLUSTER=gke_istio-security-testing_us-central1-a_istio-testing ./setup_test.sh 1.7.0 pre-release
    ```

To cleanup the previous setup with n number of deployments use the following command:

    ```bash
    NUM=100 CLUSTER=gke_istio-security-testing_us-central1-a_istio-testing ./cleanup.sh
    ```

Here n=100.

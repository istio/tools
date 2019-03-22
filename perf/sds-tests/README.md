To run the SDS test that goes to Citadel:
- Create a GKE cluster and set it as the current cluster.
You may use "kubectl config current-context" to confirm that this cluster
is set as the current cluster.

- Deploy Istio:
Let the root directory of this repo be ROOT-OF-REPO.
  cd ROOT-OF-REPO/perf/istio
Run the following command. You may replace release-1.1-20190208-09-16
in the command to the release to test.
  DNS_DOMAIN=your-example-domain VALUES=values-istio-sds-auth.yaml ./setup.sh release-1.1-20190208-09-16
After running the above script to deploy Istio, wait a moment for Istio to be ready.

- Deploy an example workload that requests for certificates through SDS.
  cd ROOT-OF-REPO/perf/sds-tests
Run the following script:
  WORKLOAD_FILE=./httpbin.yaml ./setup_test.sh
Wait a moment for the deployment to be ready. Then view the logs of Node Agents.
The Node Agents can be listed through
the command "kubectl get pod -n istio-system | grep nodeagent".
You should see the following log entries in some of the Node Agents that show
SDS pushing certificates to the Envoy of the example workload. When a certificate
issuance or rotation occurs, the following log entries are generated.
  info    SDS: push root cert from node agent to proxy
  info    SDS: push key/cert pair from node agent to proxy

- After testing, you may delete the Istio and example workload created for this test
by running:
  ./cleanup.sh

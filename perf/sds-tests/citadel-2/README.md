To run the SDS test that goes to Citadel:
- Create a GKE cluster and set it as the current cluster.
Here this test is ran on the cluster *sds-citadel-cert-rotation-2*
on GCP project *istio-security-testing*.
You may use `kubectl config current-context` to confirm that your newly created cluster
is set as the current cluster.

- Deploy Istio:
Let the root directory of this repo be *ROOT-OF-REPO*.
Run the following commands:
```
  cd ROOT-OF-REPO/perf/istio
  DNS_DOMAIN=your-example-domain VALUES=values-istio-sds-auth.yaml ./setup.sh release-1.1-20190213-09-16
```  
You may replace the Istio release
in the command to the Istio release to test.
After running the above script to deploy Istio, wait a moment for Istio to be ready.

- Deploy workloads that requests for certificates through SDS.
The following example command will deploy 3 httpbin and sleep workloads.
Note: the number of workloads can be ran depends on the size of your cluster.
```
  cd ROOT-OF-REPO/perf/sds-tests/citadel-2
  NUM=3 ./setup_test.sh
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

- Run the following commands to test mTLS by curl httpbin from sleep 
(replace the example sleep and httpbin pod names in the example command with
the pod names in your test). When the command succeeds, `HTTP/1.1 200 OK` should
be displayed.
```
  kubectl exec -it sleep-6f784fb648-7mvgp -c sleep -- curl -v httpbin:8000/headers
```

- After testing, you may delete the Istio and example workload created for this test
by running:
```
  ./cleanup.sh
```
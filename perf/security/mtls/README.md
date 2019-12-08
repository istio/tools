# Mutual TLS Performance Evaluation

This test suite measures the performance overhead due to the mutual TLS encryption, specifically
the data plane CPU, latency and throughput.

**Test Setup**

- Istio is installed with auto mTLS enabled. DestinationRule ISTIO_MUTUAL is to be deprecated for mTLS.
Relying on auto mTLS is more future proof.
- Deployment is described via service graph.
- Load client is sending traffic to two group of service in the service graph:
  1. mtls frontend -> mtls backend.
  1. plaintext frontend -> plaintext backend


## Instructions

1. Install Istio

    ```shell
    export source setup.sh && setup_istio
    ```
1. Wait Istio is ready, pods and ingress ip is assigned, and install test workloads

    ```shell
    source setup.sh && setup_test
    ```

1. Wait for couple of hours to run the test. Record the CPU/Memory footprint. Metrics we consider:
   
   1. Performance Dashboard, vCPU and memory, filtering by `namespace = mtls`, focused on istio-proxy
   1. Workload dashboard, looking at the latency and success rate.

1. Toggle mTLS status by configure policy in `mtls` namespace to be plaintext. Auto mTLS is assumed,
deleting policy is all we need to do.

    ```shell
    kubectl apply -f plaintext.yaml -n mtls
    ```

## Data

Following is the test results, with Istio release SHA at `438827b1602037fe30dedcd0008ce0cae0ef0aee`,
a commit at the end of the Istio 1.4 release.

`data/` folder contains the screenshot from grafana for performance data.

1. Without mTLS,
   - 
1. With mTLS enabled, 15:22, 1203-2019
   - 

## References

Here are some links of existing TLS overhead discussion.

- [IETF TLS overhead Memo](https://tools.ietf.org/id/draft-mattsson-uta-tls-overhead-01.html#rfc.section.3.3).
- [Stackoverflow](https://stackoverflow.com/questions/1615882/how-much-network-overhead-does-tls-add-compared-to-a-non-encrypted-connection)
new TLS session requries 6.5K bytes, the messsage itself is the same size.

## Misc

TODO

- stalled here, need to find some data on where is stored, png files.
- no traffic when plaintext.
- probe rewrite does not seem work with tls
- gateway and vs in other ns conflict with the test ns, how to isolate?
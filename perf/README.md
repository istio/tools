# Istio Performance/Stability Testing

This folder contains tests for performance and stability. There are different types of test under each subdirectory. For more details, see each directories README.

1. [/istio-install](./istio-install) provides scripts and Helm values to setup Istio for performance testing.

    This setup is designed for very large clusters to test Istio's limits. Most tests can run on a standard Istio install.
1. [/stability](./stability) provides tests that exercise various Istio features to ensure stability.

    The intent of these tests is to be run continuously for extend periods of time, which differentiates them from integration tests.
1. [/benchmark](./benchmark) provides a test to measure the latency and metrics of traffic between pods in various setups.
1. [/load](./load) provides tools to generate large services to test Istio under heavy load.
1. [/meshery](https://github.com/layer5io/meshery) provides benchmarking and management for Istio.

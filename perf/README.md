# Istio Performance/Stability Testing

This folder contains tests for performance and stability. There are different types of test under each subdirectory. For more details, see each directories README.

1. [/istio-install](./istio-install) provides scripts and Helm values to setup Istio for performance testing.

    This setup is designed for very large clusters to test Istio's limits. Most tests can run on a standard Istio install.
1. [/stability](./stability) provides tests that exercise various Istio features to ensure stability.

    The intent of these tests is to be run continuously for extend periods of time, which differentiates them from integration tests.
1. [/benchmark](./benchmark) provides a test to measure the latency and metrics of traffic between pods in various setups.
1. [/load](./load) provides tools to generate large services to test Istio under heavy load.


## Setup Istio

For performance testing, it is recommended to setup Istio with performance oriented values, but it is not required.

### Setup With Performance Parameters

Look at values.yaml for details.

To setup Istio, run `DNS_DOMAIN=your-example-domain ./setup_istio.sh release-1.1-20190125-09-16`.

To just output the deployment file, run `DRY_RUN=1 DNS_DOMAIN=your-example-domain ./setup.sh release-1.1-20190125-09-16`.

You may replace the release in the command to the release to test.

You may also override the Helm repo or release URL:

```bash
export HELMREPO_URL=https://storage.googleapis.com/istio-release/releases/1.1.0-rc.0/charts/index.yaml
export RELEASE_URL=https://github.com/istio/istio/releases/download/untagged-c41cff3404b8cc79a97e/istio-1.1.0-rc.0-linux.tar.gz

DNS_DOMAIN=your-example-domain ./setup.sh release-1.1-20190203-09-16
```

### Setup With Custom Installation

The performance tests can also run on an existing Istio installation. This can be useful to test out settings or modifications to Istio.

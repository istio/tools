# Istio Performance/Stability Testing

The intent of these tests is to create a cluster running various different services, with different setups and patterns, to help understand how Istio performs in a real-world scenario.

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

## Setup Tests

To run the tests, run `./setup_tests.sh setup`. To delete them, run `./setup_tests.sh delete`.

You can also set env variables:
* `NAMESPACE` to specify a custom namespace for tests
* `DRY_RUN` to just generate the yaml files without applying
* `TESTS` space separate list of tests to run. Example: `TESTS="http10 graceful-shutdown"`

### Default Tests
For details on the tests, read the README in each directory.

* http10
* graceful-shutdown
* gateway-bouncer

### Optional Tests

* allconfig - currently has some bugs
* sds-certmanager - requires gcloud to configure GCP DNS, and a gcp DNS zone set as env variable DNS_ZONE

## Deleting Tests

To delete all installed tests, run `./setup_tests.sh delete`.

If you provided any options when setting up the tests, such as `NAMESPACE` and `TESTS`, you will need to specify these again to ensure all tests are deleted.

## Analyzing Performance

The Grafana dashboards are a useful tool to analyze the performance and health of the tests.

In addition, the [metrics tool](/metrics/check_metrics.py) can pull metrics from Prometheus and analyze them to determine the health of each scenario. This is especially useful for some tests, where it is unclear what "good behavior" looks like just from looking at Grafana.

## Adding Tests

To add a new scenario, create a new folder with a Helm chart that sets up your scenario.

To add it as a default test, add it to the `ALL_TESTS` variable in [setup_tests.sh](/perf/istio/setup_tests.sh).

If your scenario involves configuration beyond what is possible by Helm, you may need to add custom logic in the `setup_tests` function of [setup_tests.sh](/perf/istio/setup_tests.sh).

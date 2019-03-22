# Stability Tests

This directory providestests that exercise various Istio features to ensure stability.
                           
The intent of these tests is to be run continuously for extend periods of time, to ensure features are stable over long periods of time and in real world scenarios, which differentiates them from integration tests.


## Setup Tests

To run the tests, run `./setup_tests.sh setup`. To delete them, run `./setup_tests.sh delete`.

You can also set env variables:
* `NAMESPACE` to specify a custom namespace for tests
* `DRY_RUN` to just generate the yaml files without applying
* `TESTS` space separate list of tests to run. Example: `TESTS="http10 graceful-shutdown"`

### Default Tests

For details on the tests, read the README in each directory.

* http10 - tests http 1.0 support
* graceful-shutdown - tests graceful termination of connections when services are terminated.
* gateway-bouncer - tests gateway readiness features 

With Istio defaults, these test will require around 2 vCPUs and 2GB of memory.
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

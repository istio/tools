# Stability Tests

This directory providestests that exercise various Istio features to ensure stability.

The intent of these tests is to be run continuously for extend periods of time, to ensure features are stable over long periods of time and in real world scenarios, which differentiates them from integration tests.

## Setup Tests

To run the tests, run `make stability`. To delete them, run `make clean-stability`.

You can also set env variables:
* `DRY_RUN` to just generate the yaml files without applying

To run only some tests, run `make TEST`. For example, `make mysql`.

### Default Tests

For details on the tests, read the README in each directory.

* http10 - tests http 1.0 support
* graceful-shutdown - tests graceful termination of connections when services are terminated.
* gateway-bouncer - tests gateway readiness features
* redis - installs a Redis setup with master, slave, and client
* mysql

With Istio defaults, these test will require around 2 vCPUs and 2GB of memory.

### Optional Tests

These tests are not enabled by default, but can be run individually or with `make stability_all`.

* istio-chaos-partial - disabled by default, as impacts the entire Istio install by killing all but one instance of an Istio component (or the single instance if there is only one).
* istio-chaos-total - disabled by default, as impacts the entire Istio install by scaling Istio components to zero.
* istio-upgrader - disabled by default, as impacts the entire Istio install by redeploying Istio components.
* allconfig - currently has some bugs
* sds-certmanager - requires gcloud to configure GCP DNS, and a gcp DNS zone set as env variable DNS_ZONE

## Deleting Tests

To delete all installed tests, run `make clean-stability`.

If you provided any options when setting up the tests, such as `NAMESPACE` and `TESTS`, you will need to specify these again to ensure all tests are deleted.

## Analyzing Performance

The Grafana dashboards are a useful tool to analyze the performance and health of the tests.

In addition, the [metrics tool](/metrics/check_metrics.py) can pull metrics from Prometheus and analyze them to determine the health of each scenario. This is especially useful for some tests, where it is unclear what "good behavior" looks like just from looking at Grafana.

## Adding Tests

To add a new scenario, create a new folder with a Helm chart that sets up your scenario, then add it to `stability.mk`.

# Stability Tests

This directory providestests that exercise various Istio features to ensure stability.

The intent of these tests is to be run continuously for extend periods of time, to ensure features are stable over long periods of time and in real world scenarios, which differentiates them from integration tests.

## Release Qualification Test

### Introduction

The long running test script would deploy service graphs application with 15 namespaces in the cluster and run continuously. Prometheus/Alertmanager related resources, monitors and alerting rules would be deployed and managed by Prometheus Operator, Grafana would be installed via sample addon config yaml.

There is a webhook pod deployed which handles the alertmanager alerts and notification:

1. Abnormal metrics breaking SLO would be recorded in the alertmanager-webhook pod
1. For the pod to write monitor status to the spanner tables, you have to
    1. create the cluster with full access to Google Cloud API or configure the cluster with Workload Identity
    1. create the spanner tables first in your own project or reuse the existing one in istio-testing
1. Optionally, corresponding alertmanager notification can be pushed to slack channel, checkout the example config for [slack webhook](./alertmanager/values.yaml#L21). Suspicious logs would be scanned and recorded in the istio-logs-checker Cronjob.

### Run the script

If you want to run against a public release(stable or dev), specify the target release TAG/VERSION/RELEASE_URL and you can pass extra arguments to istioctl install, check more details about accepted argument at [install_readme] (../istio-install#setup-istio). You can specify the namespace number of the servicegraph workloads by setting NAMESPACE_NUM var. Set the DNS_DOMAIN so that prometheus/grafana can be exposed and accessed.

For example

run with Istio 1.9.0:

`DNS_DOMAIN=release-qual-19.qualistio.org VERSION=1.9.0 NAMESPACE_NUM=15 ./long_running.sh`

run with Istio 1.7.1 prerelease:

`DNS_DOMAIN=release-qual-19.qualistio.org VERSION=1.9.1 NAMESPACE_NUM=15  ./long_running.sh --set hub=gcr.io/istio-prerelease-testing --set tag=1.9.1`

run with Istio private release:

`RELEASE_URL=gs://istio-private-prerelease/prerelease/1.9.3/istio-1.9.3-linux.tar.gz NAMESPACE_NUM=15  ./long_running.sh --set hub=gcr.io/istio-prow-build --set tag=1.9.3`

You can also run the canary upgrade mode, which would deploy an extra cronjob to automatically upgrade the control plane and data plane to the latest dev release every 48h, default is off.

`CANARY_UPGRADE_MODE=true DNS_DOMAIN=release-qual-19.qualistio.org VERSION=1.9.0 NAMESPACE_NUM=15 ./long_running.sh`

If you already install specific Istio version in the cluster, you can also point to the local release bundles to install grafana, some extra prometheus configs needed, e.g.

`LOCAL_ISTIO_PATH=/Users/iamwen/Downloads/istio-1.9.3 ./long_running.sh`

Or if you want to skip Istio setup completely, just deploy the workloads and alertmanager

`VERSION=1.6.5 NAMESPACE_NUM=15 SKIP_ISTIO_SETUP=true ./long_running.sh`

Note:
It is likely the script would fail in between because of transient issues such as node rescaling as more workloads being deployed. Just rerun the script again accordingly when that happens, for example it is likely that the scaling happen at the stage of deploying workload, you can just rerun from the failing namespace like:
`VERSION=1.6.5 NAMESPACE_NUM=15 SKIP_ISTIO_SETUP=true ./long_running.sh --set hub=gcr.io/istio-prerelease-testing --set tag=1.6.5`

### Monitor List

The monitors are configured via PrometheusRule CR managed by prometheus operator. Check the [list of provided monitors](./alertmanager/prometheusrule.yaml) and update correspondingly based on your requirements.

### Dashboard

1. The alertmanager webhook pod would write the monitor status data to two spanner table: MonitorStatus and ReleaseQualTestMetadata

1. The [eng.istio.io](http://eng.istio.io/releasequal) would read from the spanner table instances in GCP istio-testing project

For release managers, to reuse the existing spanner instance and publish the result to [eng.istio.io](http://eng.istio.io/releasequal), run the test on a new cluster under istio-testing GCP project and make sure the PROJECT_ID is set to istio-testing

The related params of spanner table are defined in env variables of the [webhook deployment](./alertmanager/templates/alertmanager-webhook.yaml)

### Checking Result

Unlike the normal integration/unit test that we can declare success/failure directly, it is up to release managers to make decision based on the metrics collected during the test.

Important metrics: Success rate, Latency, CPU/Memory of Istiod and Proxies

Normally the test would be running for at least 48h and here are the steps that we can check intermittently during the test and after 48h:

1. Check the control plane and workload basic status

1. Check the prometheus monitor alert status via Prometheus UI or Spanner Table(If running in istio-testing, check the [eng.istio.io dashboard](http://eng.istio.io/releasequal))

1. Check the Grafana dashboard for more detailed metric over time

1. Check the cronjob pod in the logs-checker namespace, grep for `found suspicious logs from svc`

### Upgrade

There is an optional cronjob deployed to do a canary upgrade to the latest dev build of specific branch, to enable that just
set environment variable `CANARY_UPGRADE_MODE` to `true` before running [long_running.sh](./long_running.sh)

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

In addition, the [metrics tool](../../metrics/check_metrics.py) can pull metrics from Prometheus and analyze them to determine the health of each scenario. This is especially useful for some tests, where it is unclear what "good behavior" looks like just from looking at Grafana.

## Adding Tests

To add a new scenario, create a new folder with a Helm chart that sets up your scenario, then add it to `stability.mk`.

# Stability Tests

This directory providestests that exercise various Istio features to ensure stability.

The intent of these tests is to be run continuously for extend periods of time, to ensure features are stable over long periods of time and in real world scenarios, which differentiates them from integration tests.

## Run Release Qualification Test

The long running test would deploy service graphs application with 15 namespaces in the cluster and run continuously. Prometheus/Alertmanager would be installed and managed by Prometheus Operator, Grafana would be installed via sample addon config yaml.

Abnormal metrics breaking SLO and suspicious logs would be recorded in the alertmanager-webhook pod, corresponding notification would be pushed to slack channel #stability-test(Note: any slack webhook url uploaded to github would be deactivated, contact @richardwxn for latest webhook url to put in the alertmanagerconfig.yaml)

If you want to run against a public release(stable or dev), specify the target release TAG/VERSION/RELEASE_URL and you can pass extra arguments to istioctl install, check more details about accepted argument at [install_readme](https://github.com/istio/tools/tree/master/perf/istio-install#setup-istio). You can specify the namespace number of the servicegraph workloads by setting NAMESPACE_NUM var.

For example

run with istio 1.5.4:

`VERSION=1.5.4 NAMESPACE_NUM=15 ./long_running.sh`

run with istio 1.6.5 prerelease:

`VERSION=1.6.5 NAMESPACE_NUM=15  ./long_running.sh --set hub=gcr.io/istio-prerelease-testing --set tag=1.6.5`

If you already install specific Istio version in the cluster, you can also point to the local release bundles to install grafana, some extra prometheus configs needed, e.g.

`LOCAL_ISTIO_PATH=/Users/iamwen/Downloads/istio-1.5.5 ./long_running.sh`

Or you skip Istio setup completely, just deploy the workloads and alertmanager

`VERSION=1.6.5 NAMESPACE_NUM=15 SKIP_ISTIO_SETUP=true ./long_running.sh`

Note:
1. This does not support running with asm managed control plane yet.
2. It is likely the script would fail in between because of transient issues such as node rescaling as more workloads being deployed. Just rerun the script again accordingly when that happens, for example it is likely that the scaling happen at the stage of deploying workload, you can just rerun from the failing namespace like:
`VERSION=1.6.5 NAMESPACE_NUM=15 SKIP_ISTIO_SETUP=true ./long_running.sh --set hub=gcr.io/istio-prerelease-testing --set tag=1.6.5`

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

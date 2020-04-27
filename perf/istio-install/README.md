# Istio Install

This folder provides tools to setup a cluster with Istio for performance testing.

# GCP Cluster Setup

Required environment:

```shell

PROJECT_ID - GCP project id, for example istio-testing
CLUSTER_NAME - name of the cluster to setup, for example istio14test1
ZONE - zone where the cluster will be setup, for example us-central1-a
REGION - region where a regional cluster will be set up, for example us-central1. Overrides ZONE.
DNS_DOMAIN - domain to use for TLS cert testing.

Optional:
export MACHINE_TYPE=n1-standard-4 - will use a small machine, for testing stability in small clusters.
export IMAGE=UBUNTU - will use ubuntu instead of the recommended COS
export MIN_NODES=1 - will start with 1 instead of default 4
export ISTIO_VERSION - installed version of istio, will be set as a label on nodes

```

For load testing, the setup requires a very large cluster - at least 32 vCPUs reserved for Istio is recommended.
The defaults values are 32vCP and at least 4 nodes.

For testing stability and e2e behavior in small clusters - 4vCPU per node and 1 node with auto-scaling should work.

The script will create files to be used later in the setup, as config maps:
- `${CLUSTER_NAME}/google-cloud-key.json` - will be used for authenticating control plane for GCP operations.
- `${CLUSTER_NAME}/kube.yaml` - credentials for accessing k8s
- `${CLUSTER_NAME}/configmap*` - configmaps with GCP-specific configurations

# Setup Istio

The `setup_istio.sh` scripts is a helper to install Istio with specific configurations for performance testing. The script
provides a few ways to specify which version to install:

* `TAG`: for example `1.6-dev`. This will download the latest [dev build](https://github.com/istio/istio/wiki/Dev%20Builds) for the tag.
* `DEV_VERSION`: for example, `1.4-alpha.41dee99277dbed4bfb3174dd0448ea941cf117fd`. This will download the specific [dev build](https://github.com/istio/istio/wiki/Dev%20Builds).
* `VERSION`: for example, `1.2.3`. This will download a specific release version specified.
* `RELEASE_URL`: for example, `https://example.com/istio.tar.gz`. This will download an arbitrary tar.gz.
* `GCS_URL`: for example, `gs://example/istio.tar.gz`. Same as `RELEASE_URL`, but will use `gsutil` to download.

Architecture will be automatically detected, but can be overrided. For example, `ARCH_SUFFIX=linux`.

In addition to setting up the core Istio, the prometheus operator and gateways for the telemetry addons will be setup. Pass `SKIP_EXTRAS` to skip these.

Arguments to the script will be passed to `istioctl during install`. For example, to install the latest version with the default config file:

```
TAG=latest ./setup_istio.sh -f istioctl_profiles/default.yaml
```

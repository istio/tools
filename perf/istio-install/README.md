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
- ${CLUSTER_NAME}/google-cloud-key.json - will be used for authenticating control plane for GCP operations.
- ${CLUSTER_NAME}/kube.yaml - credentials for accessing k8s
- ${CLUSTER_NAME}/configmap* - configmaps with GCP-specific configurations

## Setup Istio with different releases

Below commands will setup istio with corresponding release using [default performance testing overlay file](https://github.com/istio/tools/blob/master/perf/istio-install/istioctl_profiles/default.yaml)

1. To setup Istio with a specific release, run `DNS_DOMAIN=your-example-domain ./setup_istio_release.sh 1.5.0`

1. To setup Istio with latest build of a dev stream, run `DNS_DOMAIN=your-example-domain ./setup_istio_release.sh 1.5-dev`.

1. To setup Istio with prerelease candidates, run `DNS_DOMAIN=your-example-domain ./setup_istio_release.sh 1.5.0-alpha.0 pre-release`

1. To replace default overlay with other [predefined overlay files](https://github.com/istio/tools/blob/master/perf/istio-install/istioctl_profiles),
run `export CR_FILENAME="automtls.yaml" && DNS_DOMAIN=your-example-domain ./setup_istio_release.sh 1.5.0`

1. To provide extra override via set or self defined yaml file, run with EXTRA_ARGS, for example: `EXTRA_ARGS="--set values.grafana.enabled=true -f overlay.yaml" DNS_DOMAIN=your-example-domain ./setup_istio_release.sh 1.5.0`
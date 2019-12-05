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

## Setup with operator

Since Istio 1.4, operator provided by `istioctl` becomes the default installation mechanism.
`setup_istio_operator.sh` provides the automation. You can add your own operator profile, and then
setup Istio installation via running the script. For example,

```shell
export OPERATOR_PROFILE="automtls.yaml" && ./setup_istio_operator.sh
```

## Setup With Performance Parameters

Look at values.yaml for details on the parameters.

To setup Istio with a specific release, run `DNS_DOMAIN=your-example-domain ./setup_istio.sh release-1.1-20190125-09-16`.

To setup Istio with latest of a branch, run `DNS_DOMAIN=your-example-domain ./setup_istio.sh release-1.2-latest`.
This command will setup the latest build from the 1.2 release branch.

To just output the deployment file, run `DRY_RUN=1 DNS_DOMAIN=your-example-domain ./setup_istio.sh release-1.1-20190125-09-16`.

### Latest release

DNS_DOMAIN=v11p.qualistio.org ./setup_istio.sh release-1.1-latest
You may replace the release in the command to the release to test.

You may also override the Helm repo or release URL:

```bash
export HELMREPO_URL=https://storage.googleapis.com/istio-release/releases/1.1.0-rc.0/charts/index.yaml
export RELEASE_URL=https://github.com/istio/istio/releases/download/untagged-c41cff3404b8cc79a97e/istio-1.1.0-rc.0-linux.tar.gz

DNS_DOMAIN=your-example-domain ./setup_istio.sh release-1.1-20190203-09-16
```

### Overwrite helm flags

To overwrite helm flags, create a file to hold helm flags you want to overwrite and save as extra-values.yaml or other file names.

```bash
DNS_DOMAIN=your-example-domain EXTRA_VALUES=extra-values.yaml ./setup_istio.sh release-1.1-20190203-09-16
```

### Installing release candidates

```bash
DNS_DOMAIN=v11mcp.qualistio.org ./setup_istio_release.sh 1.1.4 pre-release
```

This option uses pre-release buckets to get builds and charts.

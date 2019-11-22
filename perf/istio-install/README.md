# Istio Setup

For performance testing, it is recommended to setup Istio with performance oriented values, but it is not required.

This setup requires a very large cluster - at least 32 vCPUs reserved for Istio is recommended.

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

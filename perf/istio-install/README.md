# Istio Setup

For performance testing, it is recommended to setup Istio with performance oriented values, but it is not required.

This setup requires a very large cluster - at least 32 vCPUs reserved for Istio is recommended.

## Setup Istio with different releases

To setup istio, run `./setup_istio_releash.sh TAG RELEASE_TYPE`

1. To setup Istio with a specific release, run `DNS_DOMAIN=your-example-domain ./setup_istio_release.sh 1.3.3 release`.

1. To setup Istio with latest build of a dev release, get the dev release tag first from https://gcsweb.istio.io/gcs/istio-build/dev/, then run `DNS_DOMAIN=your-example-domain ./setup_istio_release.sh 1.4-alpha.0039742337ddf3b766ac974e9a7aad003896cfcf dev`.

1. To setup Istio with prerelease candidate, run ```DNS_DOMAIN=v11mcp.qualistio.org ./setup_istio_release.sh 1.4.0-alpha.0 pre-release
                                              ```

1. To just output the deployment file, run `DRY_RUN=1 DNS_DOMAIN=your-example-domain ./setup_istio_release.sh 1.3.3 release`.


You may also override the Helm repo or release URL and run ./setup_istio.sh directly, for example:

```bash
export HELMREPO_URL=https://storage.googleapis.com/istio-release/releases/1.3.3/charts/index.yaml
export RELEASE_URL=https://github.com/istio/istio/releases/download/1.3.3/istio-1.3.3-linux.tar.gz

DNS_DOMAIN=your-example-domain ./setup_istio.sh 1.3.3
```

### Overwrite helm flags
Look at values.yaml for details on the parameters. 

To overwrite helm flags, create a file to hold helm flags you want to overwrite and save as extra-values.yaml or other file names.

```bash
DNS_DOMAIN=your-example-domain EXTRA_VALUES=extra-values.yaml ./setup_istio_release.sh 1.3.3 release
```
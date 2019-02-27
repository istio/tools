# Istio Performance Testing

## Setup Istio With Performance Parameters

For performance testing, it is recommended to setup Istio with performance oriented values. Look at values.yaml for details.

To setup Istio, run `DNS_DOMAIN=your-example-domain ./setup_istio.sh release-1.1-20190125-09-16`.

To just output the deployment file, run `DRY_RUN=1 DNS_DOMAIN=your-example-domain ./setup.sh release-1.1-20190125-09-16`.

You may replace the release in the command to the release to test.

### Override RELEASE and HELMREPO URLS

```bash
export HELMREPO_URL=https://storage.googleapis.com/istio-release/releases/1.1.0-rc.0/charts/index.yaml
export RELEASE_URL=https://github.com/istio/istio/releases/download/untagged-c41cff3404b8cc79a97e/istio-1.1.0-rc.0-linux.tar.gz

DNS_DOMAIN=your-example-domain ./setup.sh release-1.1-20190203-09-16
```

## Setup Tests

To run the tests, run `./setup_tests.sh setup`. To delete them, run `./setup_tests.sh delete`.

You can also set env variables:
* `NAMESPACE` to specify a custom namespace for tests
* `DRY_RUN` to just generate the yaml files without applying
* `TESTS` space separate list of tests to run. Example: `TESTS="http10 graceful-shutdown"`

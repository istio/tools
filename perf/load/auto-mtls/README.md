# Auto Mutual TLS Test

Istio auto mutual TLS is a feature to automatically configure mutual TLS between Envoy sidecar.

This feature can be affected by

- The destination service having Envoy sidecar or not.
- `AuthenticationPolicy`, `DestinationRule` configuration.

Thus we setup the test to simulate traffic meanwhile updating deployments with or without Envoy
sidecar simutaneously.

A service graph instance with 5 workloads, service 0, 1, 2 calls service 3 and automtls.

- Service `automtls` has workloads with and without sidecar in mixed mode.
- All other services workloads instances have sidecar injected.

TODO(incfly):

1. Add script to update the authn policy.
2. Generate load using load client, plus required virtual service.
3. Verify the grafana dashboard load.

## Steps to Run Test

1. Install Istio:

```bash
pushd ../../istio-install
export ISTIO_RELEASE="release-1.2-latest"  # or any Istio release
export DNS_DOMAIN=local
export EXTRA_VALUES=values-auto-mtls.yaml
./setup_istio.sh $ISTIO_RELEASE
popd
```

1. Setup the Tests

```bash
./setup_test.py
# TODO, figure it out.
krm meshpolicy default
```

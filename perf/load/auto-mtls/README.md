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
1. Verify the grafana dashboard load.
1. Optional, emit metrics to correlate the config chang with traffic stability on Prometheus.

## Steps to Run Test

1. Install Istio:

```bash
export ISTIO_RELEASE="1.4-alpha.bc8b1ebdffacd65d77365597ff73811346f3f11c"  # or any Istio release
export DNS_DOMAIN=local
export EXTRA_VALUES=values-auto-mtls.yaml
# Install istio
./istio.sh
```

1. Setup the Tests

```bash
# Setup the test.
./setup.sh
```


Issue:

- Installer does not work with the grafana
- Do not see `istioctl manifest`.
- Don't want library dependency in istio/tools libary.
- rewrite the installer and install from istioctl myself...
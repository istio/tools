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

## Steps to Run Test

1. Install Istio:

```bash
export ENABLE_AUTO_MTLS=true
./istio.sh
```

1. Setup the Tests

```bash
# Setup the test.
./setup.sh
```

1. Clean Up

```bash
kubectl rm ns istio-system automtls
```

1. Run base line for control

```bash
export ENABLE_AUTO_MTLS=false
./istio.sh
```

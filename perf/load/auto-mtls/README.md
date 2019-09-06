# Auto Mutual TLS Test

Istio auto mutual TLS is a feature to automatically configure mutual TLS between Envoy sidecar.

This feature can be affected by

- The destination service having Envoy sidecar or not.
- `AuthenticationPolicy`, `DestinationRule` configuration.

Thus we setup the test to simulate traffic meanwhile updating deployments with or without Envoy
sidecar simutaneously.

In order to run the test

```bash
./setup_test.py
```

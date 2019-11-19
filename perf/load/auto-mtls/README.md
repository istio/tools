# Auto Mutual TLS Test

Istio auto mutual TLS is a feature to automatically configure mutual TLS between Envoy sidecar.

This feature can be affected by

- The destination service having Envoy sidecar or not.
- `AuthenticationPolicy`, `DestinationRule` configuration.

Thus we setup the test to simulate traffic meanwhile updating deployments with or without Envoy
sidecar simutaneously.

The service graph consists of two groups of services

- `svc-0-front` is the client, all with sidecars
- `svc-0-back-partial-istio`, means backend, with istio sidecar injected, `*-legacy` means no sidecar.

## Steps to Run Test

1. Install Istio:

```bash
export ENABLE_AUTO_MTLS=true
./istio.sh
```

1. Setup the Tests

```bash
# sleep just waiting to ensure ingress gateway IP is finalized, needed for fortio client.
sleep 30
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

1. Gather Metrics

```bash
pid=$(kpidn istio-system -lapp=grafana)
kpfn istio-system $pid 3000
```

Select workload dashboard, focused on `svc-0-front`, extend time range correspondingly.

- `Outgoing duration`
- `Outgoing success rate`

Open `Share` button, click through `https://snapshot.raintank.io/`, ensures setting a longer timeout
to grab all the metrics.

# Auto Mutual TLS Test

Istio auto mutual TLS is a feature to automatically configure mutual TLS between Envoy sidecar.

Since this feature can be affected by whether the destination service having Envoy sidecar or not,
we setup the test to simulate traffic meanwhile updating deployments with or without Envoy
sidecar simutaneously.

The service graph consists of two groups of services

- `svc-0-front` is the client, all with sidecars, calling `svc-0-back` service.
- `svc-0-back-istio`, means backend, with istio sidecar injected, `*-legacy` means no sidecar.

## Steps to Run Test

1. Install Istio:

```bash
pushd ../../istio-install
export OPERATOR_PROFILE="automtls.yaml" && ./setup_istio_operator.sh
popd
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

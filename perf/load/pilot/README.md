# Pilot Load Test

## Basic Setup

This folder includes a chart for light weight pilot performance measurements. Each workload only has
one Envoy sidecar, single container in a Pod.

In 1.1.2 release, as resource consumption data point,
Envoy sidecar consumes 5m CPU and 128M memory.

## Config Push Latency Test

Config push measures the latency of the CDS push, the time starting from service discovery
changes to the corresponding CDS changes reflected in the Envoy config.

Several special installation setup is needed.

- Remove Prometheus metrics drop in the config map
- Ensure the cluster is appropriated provisioned with enough memory

`./load_test.py`, starts large deployment, measure how long it takes for all
sidecar to receive the same cds version.

## Setup

To support large performance testing, many things need to be specially tuned.

- Install Istio version, e.g., 1.1.2.
- Pilot resource consumption, 6GB memory, 5 replica.
- Prometheus resource consumption, update to 0.5 cpu, 10GB memory.
- Prometheus filtering.

  ```yaml
  source_labels: [ cluster_name ]
    regex: '(outbound.*svc-[1-9]+.*pilot-load|inbound|prometheus_stats).*'
    action: drop
  ```

- Run `./load-test.py`, this polling the stats from Prometheus and prints out the number of the
workloads seeing "svc-0" in their outbound CDS.
- In a seperate terminal, delete svc-0, `kubectl delete  deployment/svc-0 svc/svc-0 -npilot-load`.
- Observe the time elapsed for the `svc-0` to disappear.

## Notes and TODO

- Ensure prometheus scrape interval is set correctly. Default is 15s.
- We don't use `envoy_cluster_manager_cds_version` for now because of an [issue](https://github.com/istio/istio/issues/13994).
Different clusters does not show the same HASH even after long enough time to converge.
- Consider to use annotation
  `"sidecar.istio.io/statsInclusionPrefixes": "TBD", "gcr.io/mixologist-142215/proxyv2:suffix4"`


# Pilot Load Test

## Basic Setup

This folder includes a chart for light weight pilot performance measurements. Each workload only has one Envoy sidecar, single
container in a Pod.

In 1.1.2 release, as resource consumption data point,
Envoy sidecar consumes 5m CPU and 128M memory.

## Config Push Latency Test

`./load_test.py`, starts large deployment, measure how long it takes for all
sidecar to receive the same cds version.

TODO
- ensure prometheus scrape interval is set correctly
- `envoy_cluster_manager_cds_version` does not work well when sidecar resource is used.
- Add annotation for emiting the specific clusters, "sidecar.istio.io/statsInclusionPrefixes": "TBD"
- Using new proxy image: gcr.io/mixologist-142215/proxyv2:suffix4
- Removing Promethues load report drop config
  - `regex: '(outbound|inbound|prometheus_stats).*'`,
  - `regex: 'envoy_cluster_(lb|retry|bind|internal|max|original).*'`
- Change to `envoy_cluster_manager_cds_version{namespace="pilot-load"}` focusing on the testing namespace.
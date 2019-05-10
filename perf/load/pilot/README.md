# Pilot Load Test

## Basic Setup

This folder includes a chart for light weight pilot performance measurements.

## Config Push Latency Test

`./load_test.py`, starts large deployment, measure how long it takes for all
sidecar to receive the same cds version.

TODO
- ensure prometheus scrape interval is set correctly
- `envoy_cluster_manager_cds_version` does not work well when sidecar resource is used.
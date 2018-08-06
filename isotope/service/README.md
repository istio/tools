# Service

This directory holds the "mock-service" component for isotope. It is a
relatively simple HTTP server which follows instructions from a YAML file and
exposes Prometheus metrics.

## Usage

1. Include the entire topology YAML in `/etc/config/service-graph.yaml`
1. Set the environment variable, `SERVICE_NAME`, to the name of the service
   from the topology YAML that this service should emulate

## Metrics

Captures the following metrics for a Prometheus endpoint:

- `service_incoming_requests_total` - a counter of requests received by this
  service
- `service_outgoing_requests_total` - a counter of requests sent to other
  services
- `service_outgoing_request_size` - a histogram of sizes of requests sent to
  other services
- `service_request_duration_seconds` - a histogram of durations from "request
  received" to "response sent"
- `service_response_size` - a histogram of sizes of responses sent from this
  service

## Performance

Running on a GKE cluster with a limit of 1 vCPU and 3.75 gigabytes of memory,
and logging set to INFO, this service can reach a maximum QPS of 12,000.

# Benchmarking

This test measures performance between two pods in various set ups.

It does so with a wide variety of payload sizes and connections.

## Setup

Before running the tests, create the pods needed. The `DNS_DOMAIN` can be `local`.

```bash
export NAMESPACE=twopods
kubectl create namespace $NAMESPACE
kubectl label namespace $NAMESPACE istio-injection=enabled
DNS_DOMAIN=your_domain ./setup_test.sh
```

## Running the benchmark

Next, run the benchmark. See [runner.py](./runner/runner.py) for all options.

```bash
python runner/runner.py 16,64 1000,4000 180 --serversidecar --baseline
```

This will run the test with 16 and 64 connections, with 1000 and 4000 qps, for 180 seconds, and will test sidecar -> sidecar (on by default), client -> sidecar (`serversidecar`), and client -> server (`baseline`).

Note that the test will run all combinations of the parameters given, so this example would run 12 tests, for 3 minutes each.

## Analyzing Results

Once the tests are complete, we can extract the results from Fortio and Prometheus.

```bash
python ./runner/fortio.py FORTIO_CLIENT_URL PROMETHEUS_URL --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy
```

This will output two files, one in `json` format and another in `csv`, which holds metrics such as QPS attained, latency, and CPU/Memory metrics.

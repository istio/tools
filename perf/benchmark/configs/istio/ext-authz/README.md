# Ext-authz Benchmarking

This directory contains scripts and configurations for ext-authz benchmarking.

## Prerequisites

See [Istio Performance Benchmarking](https://github.com/istio/tools/tree/master/perf/benchmark) for environment setup.

## Run the tests

To start the tests,

```bash
./run.sh
```

The command will setup ext-authz and start running tests.

## Results

<figure>
   <img src="./results/p50.png">
   <figcaption align = "center"><b>Fig.1 - p50</b></figcaption>
</figure>

<figure>
   <img src="./results/p90.png">
   <figcaption align = "center"><b>Fig.2 - p90</b></figcaption>
</figure>

<figure>
   <img src="./results/p99.png">
   <figcaption align = "center"><b>Fig.3 - p99</b></figcaption>
</figure>

<figure>
   <img src="./results/p999.png">
   <figcaption align = "center"><b>Fig.4 - p999</b></figcaption>
</figure>

## Analysis

- The y-axis is the latency, in milliseconds; and the x-axis is the number of concurrent connections.
- We analysis 3 different loads: small (qps=100, conn=8), medium (qps=500, conn=32), and large (qps=1000, conn=64) loads.
- The latency from client to workload **with** ext-authz is longer than the latency from client to workload **without** ext-authz.
- The increase of latency to workload with ext-authz is proportional to the increase of latency to provider.

In conclusion, the extra latency that ext-authz may introduce is related to the latency to ext-authz provider, and different kinds of provider may have different results.
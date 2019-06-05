# Benchmarking

This directory contains customizable scripts to benchmark Istio's performance.

See the [Istio Performance and Scalability Guide](https://istio.io/docs/concepts/performance-and-scalability/) for performance data against the latest Istio release.  

## Prerequisites 

1. A running Kubernetes cluster with permissions to create namespaces. We recommend the following cluster specifications:

- At least 4 worker nodes
- Each node has at least 4 CPUs   

2. Latest [Istio release](https://github.com/istio/istio/releases) downloaded into this directory.  

## 1 - Setup 

1. Install the latest release of Istio. **Note**: as of Istio 1.1.7, we recommend testing with Mixer (both policy and telemetry) turned off, for improved performance.

Also note that **we do not recommend** using the [Istio quickstart install template](https://istio.io/docs/setup/kubernetes/install/kubernetes/) for any performance benchmarking, as this installation is not tuned for performance. 

Run `./install_istio.sh` to install Istio on the cluster.


2. Deploy the workloads to measure performance against.  (`DNS_DOMAIN` can be `local`.)

```bash
export NAMESPACE=twopods
kubectl create namespace $NAMESPACE
kubectl label namespace $NAMESPACE istio-injection=enabled
DNS_DOMAIN=your_domain ./setup_test.sh
```

## 2 - Run the benchmark

1. Run the benchmark, located at [runner.py](./runner/runner.py). 

```bash
python runner/runner.py 16,64 1000,4000 180 --serversidecar --baseline
```

This will run a set of [Fortio](http://fortio.org/) loadgenerator tests.

The test has 3 modes:

1) `bothsidecar` (default): measures latency where both pods have a sidecar proxy 
2) `serversidecar`: client pod has no sidecar proxy; server does 
3) `baseline`: no sidecars in either the client or server pod 

So in the example above: 16 and 64 connections, with 1000 and 4000 qps, for 180 seconds, and will test sidecar -> sidecar (on by default), client -> sidecar (`serversidecar`), and client -> server (`baseline`).

**Notes**
- `runner.py` will run all combinations of the parameters given, so this example would run 12 tests, for 3 minutes each.
- # seconds must be greater than 92 



## 3- Analyze Results

Once `runner.py` has completed, extract the results from Fortio and Prometheus.

```bash
python ./runner/fortio.py FORTIO_CLIENT_URL PROMETHEUS_URL --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy
```

`fortio.py` will output two files (one JSON, one CSV), both containing the same result metrics: Queries Per Second (QPS) attained, latency, and CPU/Memory usage. 


## 4 - Generate Charts

Use the generated CSV from `fortio.py` to : 

`graph.py <PATH_TO_CSV>` 

By default. this script outputs an `.svg` file to the current directory. Add the `--png` flag to save as a PNG image instead. 




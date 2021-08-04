# Istio Performance Benchmarking

This directory contains Python scripts to benchmark Istio's data plane performance.

See the [Istio Performance and Scalability Guide](https://istio.io/docs/concepts/performance-and-scalability/) for performance data against the latest Istio release.

For instructions on how to run these scripts with Linkerd, see the [linkerd/](linkerd/) directory.

## Prerequisites

1. [Python3](https://docs.python-guide.org/starting/installation/#installation-guides)
1. [`pipenv`](https://docs.python-guide.org/dev/virtualenvs/#virtualenvironments-ref)
1. [helm](https://helm.sh/docs/using_helm/#install-helm)

## Setup

1. Create a Kubernetes cluster.

   We provide a GKE cluster-create script in this repo, refer to [GKE Cluster Setup](../istio-install#gke-cluster-setup).

   If you are using your own cluster, see [resource requirment](../istio-install#resource-requirment) for machine type recommendations.

1. Install Istio

   See [Setup Istio](../istio-install#istio-setup).

1. Deploy the workloads to measure performance against. The test environment is two [Fortio](http://fortio.org/) pods (one client, one server), set to communicate over HTTP1, using mutual TLS authentication. By default, the client pod will make HTTP requests with a 1KB payload.

    ```bash
    export NAMESPACE=twopods-istio
    export INTERCEPTION_MODE=REDIRECT
    export ISTIO_INJECT=true
    export LOAD_GEN_TYPE=nighthawk
    export DNS_DOMAIN=v104.qualistio.org
    ./setup_test.sh
    ```

Note: We are currently supporting two types of load generators: nighthawk and fortio. Please specify the LOAD_GEN_TYPE
environment variable to your desired one.

## Prepare Python Environment

Here, `pipenv shell` will create a local Python3 virtual environment, and `pipenv install` will install all the Python
packages needed to run the benchmarking scripts and [graph_plotter](./graph_plotter) via [Pipfile](./Pipfile):

```bash
cd perf/benchmark
pipenv --three
pipenv shell
pipenv install
```

## Run performance tests

The benchmarking script is located at [runner.py](./runner/runner.py). This script runs a set of [Fortio](http://fortio.org/) or [Nighthawk](https://github.com/envoyproxy/nighthawk) performance tests depending on the kind of LOAD_GEN_TYPE you set before.

The different sidecar modes and telemetry configurations of performance tests are described in the [Istio performance dashboard](https://perf.dashboard.istio.io) site.

**How to run**:

1. run with CLI argument directly

```bash
python runner/runner.py --conn <conn> --qps <qps> --duration <duration> --OPTIONAL-FLAGS
```

1. run with config yaml

```bash
python runner/runner.py --config_file ./configs/istio/telemetryv2_stats/latency.yaml
```

Required fields to specified via CLI or config file:

- `conn` = number of concurrent connections
- `qps` = queries per second for each connection
- `duration` = number of seconds to run each test for  (the minimum value for duration should be: 92 seconds)
- `load_gen_type` = the traffic load generator type
- `telemetry_mode` = the telemetry mode you enabled while installing Istio (none or telemetryv2)

```bash
optional arguments:
  -h, --help            show this help message and exit
  --headers HEADERS     a list of `header:value` should be separated by comma, e.g. --headers="foo:bar,foo1:bar1,foo2:bar2"
  --conn CONN           number of connections, comma separated list
  --qps QPS             qps, comma separated list
  --duration DURATION   duration in seconds of the extract
  --load_gen_type LOAD_GEN_TYPE
                        traffic load generator type, can be either Fortio or Nighthawk
  --jitter JITTER       to enable or disable jitter for load generator
  --size SIZE           size of the payload
  --mesh MESH           istio or linkerd
  --telemetry_mode TELEMETRY_MODE
                        run with different telemetry configurations:
                        none, telemetryv2
  --client CLIENT       where to run the test from
  --server SERVER       pod ip of the server
  --perf PERF           also run perf and produce flame graph
  --frequency FREQUENCY sampling frequency of generating flame graph
  --ingress INGRESS     run traffic through ingress, should be a valid URL
  --extra_labels EXTRA_LABELS
                        extra labels
  --protocol_mode PROTOCOL_MODE
                        http, tcp or grpc
  --config_file CONFIG_FILE
                        config yaml file
  --cacert CACERT       path to the cacert for the fortio client inside the
                        container
  --baseline            run baseline for all
  --no_baseline         do not run baseline for all
  --serversidecar       run serversidecar-only for all
  --no_serversidecar    do not run serversidecar-only for all
  --clientsidecar       run clientsidecar-only for all
  --no_clientsidecar    do not run clientsidecar-only for all
  --bothsidecar         run both clientsidecar and serversidecar
  --no_sidecar          do not run clientsidecar and serversidecar
```

Note:
- `runner.py` will run all combinations of the parameters given. However, in order to reduce ambiguity when generating the graph, it would be
 better to change one parameter at a time and fix other parameters
- if you want to run with `--perf` flag to generate a flame graph, please make sure you have the permission to gather perf data, please refer to step 2 of this [README](https://github.com/istio/tools/tree/master/perf/benchmark/flame#setup-perf-tool)
- to test your system's maximum qps before choosing the range of qps you want to measure by passing `--qps 0`

For example:

### Example 1

```bash
python runner/runner.py --config_file ./configs/istio/telemetryv2_stats/latency.yaml
```

- This will run with configuration specified in the ./configs/istio/telemetryv2_stats/latency.yaml
- Run with telemetry v2 stats filter on and measure the latency

### Example 2

```bash
python runner/runner.py --conn 2,4,8,16,32,64 --qps 1000 --duration 240 --baseline --load_gen_type=fortio --telemetry_mode=v2-nullvm
```

- This will run separate tests for the `both` and `baseline` modes with fortio as the load generator and testing telemetryv2 enabled scenario
- Separate tests for 2 to 64 concurrent connections
- Each connection will send **1000** QPS
- Each test will run for **240** seconds

### Example 3

```bash
python runner/runner.py --conn 16,64 --qps 1000,4000 --duration 180 --serversidecar --baseline --load_gen_type=nighthawk --telemetry_mode=none
```

- 12 tests total, each for **180** seconds, with all combinations of:
- **16** and **64** connections
- **1000** and **4000** QPS
- `both`, `serversidecar`, and `baseline` modes

### Example 4

Example 1 and 2 is to gather the latency results by increasing the number of connections. If you want to gather CPU and memory related
results, you should increasing the number of QPS, like:

```bash
python runner/runner.py --conn 10  --qps 100,500,1000,2000,4000 --duration 240 --load_gen_type=fortio --telemetry_mode=v2-nullvm
```

### Example 5: CPU flame graph

```bash
python runner/runner.py --conn 10  --qps 100,500,1000,2000,4000 --duration 240 --load_gen_type=fortio --telemetry_mode=v2-nullvm --perf=true
```

This will generate corresponding `.svg` flame graph in the `perf/benchmark/flame/flameoutput` repo.
Here is the [sample output](https://github.com/istio/tools/tree/master/perf/benchmark/flame/example_flame_graph/example_output)

### Example 6

```bash
python runner/runner.py --conn 2,4,8,16,32,64 --qps 1000 --duration 240 --baseline --load_gen_type=fortio --protocol_mode=tcp
```

- This will run separate tests for the `both` and `baseline` modes with fortio tcp load test
- Separate tests for 2 to 64 concurrent connections
- Each connection will send **1000** QPS
- Each test will run for **240** seconds

## Gather Result Metrics

Once `runner.py` has completed, extract the results from Fortio and Prometheus.

1. Set `FORTIO_CLIENT_URL` to the `fortioclient` Service's `EXTERNAL_IP`:

    ```bash
    kubectl get svc -n $NAMESPACE fortioclient
    NAME           TYPE           CLUSTER-IP   EXTERNAL-IP     PORT(S)                                                       AGE
    fortioclient   LoadBalancer   xxxx          xxxx       8080:31759/TCP,8079:30495/TCP,8078:31107/TCP,8077:31034/TCP   16h

    export FORTIO_CLIENT_URL=http://$(kubectl get services -n $NAMESPACE fortioclient -o jsonpath="{.status.loadBalancer.ingress[0].ip}"):9076
    ```

    or if you don't have an external IP:

    ```bash
    kubectl -n $NAMESPACE port-forward svc/fortioclient 9076:9076 &
    export FORTIO_CLIENT_URL=http://localhost:9076
    ```

1. Set `PROMETHEUS_URL`:

    This command can be used if you installed Prometheus over the `./setup_istio.sh` script.
    Otherwise your Prometheus maybe in a different namespace and differently named.

    ```bash
    kubectl -n istio-prometheus port-forward svc/istio-prometheus 9090:9090 &
    export PROMETHEUS_URL=http://localhost:9090
    ```

1. Run `fortio.py`:

    ```bash
    python ./runner/fortio.py $FORTIO_CLIENT_URL --prometheus=$PROMETHEUS_URL --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,p999,cpu_mili_avg_istio_proxy_fortioclient,cpu_mili_avg_istio_proxy_fortioserver,cpu_mili_avg_istio_proxy_istio-ingressgateway,mem_Mi_avg_istio_proxy_fortioclient,mem_Mi_avg_istio_proxy_fortioserver,mem_Mi_avg_istio_proxy_istio-ingressgateway
    ```

    This script will generate two output files (one JSON, one CSV), both containing the same result metrics: Queries Per Second (QPS) attained, latency, and CPU/Memory usage.

## Visualize CSV Results

Please refer to [graph_plotter README](graph_plotter/README.md)

## Add new config to benchmark pipeline

Currently we are running benchmark test towards different configs as [prow job](https://prow.istio.io/?job=daily-nighthawk-performance-benchmark)

To add a new config to this pipeline, we need to add a new directory under [configs folder](https://github.com/istio/tools/tree/master/perf/benchmark/configs/istio), where we can define config parameters structured as below:

- installation.yaml: install Istio with this IstioOperator overlay file on top of istioctl built-in default profile and [perf testing default overlay](https://github.com/istio/tools/tree/master/perf/istio-install/istioctl_profiles/default-overlay.yaml)
- cpu_mem.yaml: if provided, run cpu, memory test with this config
- latency.yaml: if provided, run latency test with this config
- prerun.sh: prerun hook we want to run before test
- postrun.sh: postrun hook we want to run after test

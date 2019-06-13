# Istio Performance Benchmarking 

This directory contains Python scripts to benchmark Istio's data plane performance.

See the [Istio Performance and Scalability Guide](https://istio.io/docs/concepts/performance-and-scalability/) for performance data against the latest Istio release.  

## Prerequisites 

1. [Python3](https://docs.python-guide.org/starting/installation/#installation-guides) 
2. [`pipenv`](https://docs.python-guide.org/dev/virtualenvs/#virtualenvironments-ref) 

## Setup 

1. Create a Kubernetes cluster. We provide a GKE cluster-create script in this repo. **Note**: if using your own cluster, see the install [README](https://github.com/istio/tools/tree/master/perf/istio-install#istio-setup) for machine type recommendations. 

```bash
PROJECT_ID=<your-gcp-project>
ISTIO_VERSION=<version>
ZONE=<a-gcp-zone>
CLUSTER_NAME=<any-name>
../istio-install/create_cluster.sh $CLUSTER_NAME
```


2. Install Istio:

```bash
ISTIO_RELEASE="release-1.2-latest"  # or any Istio release
DNS_DOMAIN=local ./setup_istio.sh $ISTIO_RELEASE
```

Wait for all Istio pods to be `Running` and `Ready`:

```bash
kubectl get pods -n istio-system
```

3. Deploy the workloads to measure performance against. The test environment is two [Fortio](http://fortio.org/) pods (one client, one server), set to communicate over HTTP1, using mutual TLS authentication. By default, the client pod will make HTTP requests with a 1KB payload. 

```bash
export NAMESPACE=twopods
kubectl create namespace $NAMESPACE
kubectl label namespace $NAMESPACE istio-injection=enabled
DNS_DOMAIN=local ./setup_test.sh
```

## Prepare Python Environment 

Here, `pipenv shell` will create a local Python3 virtual environment, and `pipenv install` will install all the Python packages needed to run the benchmarking scripts (see `runner/Pipfile`). 

```
cd runner 
pipenv shell
pipenv install 
cd .. 
```

## Run performance tests 

The benchmarking script is located at [runner.py](./runner/runner.py). This script runs a set of [Fortio](http://fortio.org/) performance tests.

The test has three sidecar modes:

1) `both` (default): Client and server sidecars are present
2) `server-sidecar`: Only server sidecar is present.
3) `baseline`: Client pod directly calls the server pod, no sidecars are present.

**How to run**: 

```bash
python runner/runner.py <conn> <qps> <duration> --OPTIONAL-FLAGS
```

Where:
- `conn` = number of concurrent connections 
- `qps` = queries per second for each connection 
- `duration` = number of seconds to run each test for  (min: 92 seconds)

Optional flags:

```
  --size SIZE         size of the payload
  --client CLIENT     where to run the test from
  --server SERVER     pod ip of the server
  --perf              also run perf and produce flamegraph
  --baseline          run baseline for all
  --no-baseline       do not run baseline for all
  --serversidecar     run serversidecar-only for all
  --no-serversidecar  do not run serversidecar-only for all
  --clientsidecar     run clientsidecar and serversidecar for all
  --no-clientsidecar  do not run clientsidecar and serversidecar for all
  --ingress INGRESS   run traffic thru ingress
  --labels LABELS     extra labels
```


### Example 1 

`runner.py` will run all combinations of the parameters given. For example:


```bash
python runner/runner.py 1,2,4,8,16,32,64 1000 240 --serversidecar 
```

- This will run separate tests for the `both` and `serversidecar` modes 
- Separate tests for 1 to 64 concurrent connections 
- Each connection will send **1000** QPS 
- Each test will run for **240** seconds

### Example 2 

```bash
python runner/runner.py 16,64 1000,4000 180 --serversidecar --baseline
```

- 12 tests total, each for **180** seconds, with all combinations of: 
- **16** and **64** connections 
- **1000** and **4000** QPS 
- `both`,  `serversidecar`, and `baseline` proxy modes 


## [Optional] Disable Mixer 

Calls to Istio's Mixer component (policy and telemetry) adds latency to the sidecar proxy. To disable Istio's mixer and re-run the performance tests:


1. Disable Mixer 

```bash 
kubectl -n istio-system get cm istio -o yaml > /tmp/meshconfig.yaml
python ../../bin/update_mesh_config.py disable_mixer /tmp/meshconfig.yaml | kubectl -n istio-system apply -
```

2. Run `runner.py`, in any sidecar mode, with the `--labels=nomixer` flag.

3. Re-enable Mixer: 

```bash 
kubectl -n istio-system get cm istio -o yaml > /tmp/meshconfig.yaml
python ../../../bin/update_mesh_config.py enable_mixer /tmp/meshconfig.yaml  | kubectl -n istio-system apply -
```

## Gather Result Metrics 

Once `runner.py` has completed, extract the results from Fortio and Prometheus. 

1. Set `FORTIO_CLIENT_URL` to the `fortioclient` Service's `EXTERNAL_IP`: 

```bash
kubectl get svc -n $NAMESPACE fortioclient
```

2. Set `PROMETHEUS_URL`: 

```bash
kubectl -n istio-prometheus port-forward $(kubectl -n istio-prometheus get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') 9090:9090 & 

export PROMETHEUS_URL=http://localhost:9090 
```

3. Run `fortio.py`: 

```bash 
python ./runner/fortio.py $FORTIO_CLIENT_URL $PROMETHEUS_URL --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy
```

This script will generate two output files (one JSON, one CSV), both containing the same result metrics: Queries Per Second (QPS) attained, latency, and CPU/Memory usage. 


## Visualize Results

The `graph.py` script uses the output CSV file from `fortio.py` to generate a [Bokeh](https://bokeh.pydata.org/en/1.2.0/) interactive graph. The output format is `.html`, from which you can save a PNG image.

```bash 
runner/graph.py <PATH_TO_CSV> <METRIC>
```

Metrics options:

- `p50` (latency, 50th percentile)
- `p90` (latency, 90th percentile)
- `p99` (latency, 99th percentile)
- `mem` (server sidecar's max memory usage)
- `cpu` (server sidecar's max CPU usage)

### Example Output 

![screenshot](screenshots/bokeh-screenshot.png)
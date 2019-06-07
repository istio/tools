# Istio Performance Benchmarking 

This directory contains customizable scripts to benchmark Istio's data plane (sidecar proxy) performance.

See the [Istio Performance and Scalability Guide](https://istio.io/docs/concepts/performance-and-scalability/) for performance data against the latest Istio release.  

## Prerequisites 

1. A running Kubernetes cluster with permissions to create namespaces. We recommend the following cluster specifications:

- At least 4 worker nodes
- Each node has at least 4 CPUs   

2. Latest [Istio release](https://github.com/istio/istio/releases) downloaded into this directory.  
3. Python3 installed in your local environment. **TODO** -- document how to install dependencies for all Python scripts. 

## 1 - Setup 

1. Install Istio:

Download a release for your environment, `cd` into the Istio directory  


```bash 
kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account)

kubectl create namespace istio-system

helm template install/kubernetes/helm/istio-init --name istio-init --namespace istio-system | kubectl apply -f -

helm template ./install/kubernetes/helm/istio --name istio --namespace istio-system \
   --values ../values-istio-auth.yaml > istio.yaml

kubectl apply -f istio.yaml 
```

Wait for all Istio pods to be `Running` and `Ready`: `kubectl get pods -n istio-system`

2. Deploy the workloads to measure performance against. The test environment is two pods (one client, one server), set to communicate over HTTP, with mutual TLS. 

```bash
export NAMESPACE=twopods
kubectl create namespace $NAMESPACE
kubectl label namespace $NAMESPACE istio-injection=enabled
DNS_DOMAIN=local ./setup_test.sh
```

## 2 - Run the benchmark

1. Run the benchmark, located at [runner.py](./runner/runner.py). This will run a set of [Fortio](http://fortio.org/) loadgenerator tests.

The test has 3 modes:

1) `bothsidecar` (default): measures latency where both pods have a sidecar proxy 
2) `serversidecar`: client pod has no sidecar proxy; server does 
3) `baseline`: no sidecars in either the client or server pod 

Example: 

```bash
python runner/runner.py 10 1000 120
```

- This will run a single performance test with the default `bothsidecar` proxy mode
- Fortio will open **10** concurrent connections
- Each connection will send **1000** Queries per Second (QPS)
- The test will run for **120** seconds. (*Note* - the minimum is 92 seconds.) 

With these parameters, fortio will send a total of **12,000** requests (`# connections * QPS * seconds`). 

(Note that Fortio's raw output file is saved as JSON *inside* the `fortioclient` pod. The next step shows how to extract Fortio's results.)

### Advanced example 

`runner.py` will run all combinations of the parameters given. 

```bash
python runner/runner.py 16,64 1000,4000 180 --serversidecar --baseline
```

In this example: 

- 12 tests total, each for **180** seconds (3 minutes), with all combinations of: 
- **16** and **64** connections 
- **1000** and **4000** QPS 
- `bothsidecar` (default), `serversidecar`, and `baseline` proxy modes 


## 3- Gather Result Metrics 

Once `runner.py` has completed, extract the results from Fortio and Prometheus. 

1. Set `FORTIO_CLIENT_URL` to the `fortioclient` Service's `EXTERNAL_IP`: 

```bash
kubectl get svc -n $NAMESPACE fortioclient
```

2. Set `PROMETHEUS_URL` to Istio's Prometheus address. *note* - by default, Prometheus. 

```bash
kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') 9090:9090 & 

export PROMETHEUS_URL=http://localhost:9090 
```

Then run `fortio.py`: 

```bash 
python ./runner/fortio.py $FORTIO_CLIENT_URL $PROMETHEUS_URL --csv StartTime,ActualDuration,Labels,NumThreads,ActualQPS,p50,p90,p99,cpu_mili_avg_telemetry_mixer,cpu_mili_max_telemetry_mixer,mem_MB_max_telemetry_mixer,cpu_mili_avg_fortioserver_deployment_proxy,cpu_mili_max_fortioserver_deployment_proxy,mem_MB_max_fortioserver_deployment_proxy,cpu_mili_avg_ingressgateway_proxy,cpu_mili_max_ingressgateway_proxy,mem_MB_max_ingressgateway_proxy
```

`fortio.py` will output two files (one JSON, one CSV), both containing the same result metrics: Queries Per Second (QPS) attained, latency, and CPU/Memory usage. 

**TODO** -- document how to send results to BigQuery. 


## 4 - Visualize Results

Use the generated CSV from `fortio.py` to generate a latency vs. # connections chart. 

`graph.py <PATH_TO_CSV>`  

This script uses [Bokeh.js] to generate an interactive graph, output as an `html` file. 

By default. this script outputs an `.svg` file to the current directory. Add the `--png` flag to save as a PNG image instead. 

`graph.py` can also display CPU and Memory against # connections instead of latency. use the `--cpu` or `--memory` flag to graph that instead. 


## Linkerd Benchmarking 

In addition to a tuned Istio install, we also provide a Linkerd installation profile to run the same `../runner/` tests on. 

Sources:
- https://linkerd.io/2/getting-started/
- https://linkerd.io/2/reference/proxy-metrics/ 

See the [example-comparison](example-comparison/) directory for how to run the same tests with both Istio and Linkerd.

### 1 - create cluster 

../istio-install/create-cluster  

### 2 - install Linkerd 

./linkerd/setup-linkerd.sh <VERSION> 

### 3. deploy the fortio test environment 

```
./linkerd/setup_test.sh
``` 

### 4. Run benchmark 

```
NAMESPACE=twopods 

python runner/runner.py --linkerd 1 100 92 
```

(An example test: 1 thread, 100 QPS, for 92 seconds)

### 5. Extract Fortio latency metrics to CSV 

**Note** - Linkerd proxy CPU/memory usage not yet implemented, only latency performance.

```
export FORTIO_CLIENT_URL=<fortio svc EXTERNAL_IP:port>

python ./runner/fortio.py $FORTIO_CLIENT_URL
```

### 6. Visualize results 

```
python ./runner/graph.py --mesh=linkerd <PATH_TO_CSV> <METRIC> 
```

#### Example: 

```
python ./runner/graph.py --mesh=linkerd --xaxis=qps /var/folders/rt/610_wrfj70q8221h55lj8rvr009rz8/T/tmpn15l2dlh p90
```
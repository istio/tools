## Linkerd Benchmarking 

In addition to a tuned Istio install, we also provide a Linkerd installation profile to run the same `../runner/` tests on. 

Sources:
- https://linkerd.io/2/getting-started/
- https://linkerd.io/2/reference/proxy-metrics/ 

See the [example-comparison](example-comparison/) directory for how to run the same tests with both Istio and Linkerd.

### 1 - create cluster 

```bash 
./linkerd/istio-install/create-cluster  
```

### 2 - install Linkerd 

```bash
./linkerd/setup-linkerd.sh <VERSION> 
```

### 3. deploy the fortio test environment 

```bash
export NAMESPACE="twopods"
kubectl create namespace $NAMESPACE  
kubectl annotate namespace $NAMESPACE linkerd.io/inject=enabled
DNS_DOMAIN=local LINKERD_INJECT=enabled ./setup_test.sh
``` 

### 4. Run benchmark 

Example:

```
python runner/runner.py 16,64 1000 240 --baseline --mesh=linkerd
```

### 5. Extract Fortio latency metrics to CSV 

**Note** - Linkerd proxy CPU/memory usage not yet implemented, only latency performance.

```
export FORTIO_CLIENT_URL=<fortio client svc EXTERNAL_IP:port>

python ./runner/fortio.py $FORTIO_CLIENT_URL
```

### 6. Visualize results 

```
python ./runner/graph.py <PATH_TO_CSV> <METRIC> --mesh=linkerd 
```

#### Examples: 

**Latency, 90th percentile** 

```
python ./runner/graph.py linkerd.csv p90 --mesh=linkerd 
```

![example-linkerd-p90](linkerd-p90.png)

**Latency, 50th percentile** 

```
python ./runner/graph.py linkerd.csv p50 --mesh=linkerd 
```

![example-linkerd-p50](linkerd-p50.png)
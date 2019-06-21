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

Example:

```
python runner/runner.py 1,2,4,8,16,32,64 1000 240 --baseline --mesh=linkerd
```

### 5. Extract Fortio latency metrics to CSV 

**Note** - Linkerd proxy CPU/memory usage not yet implemented, only latency performance.

```
export FORTIO_CLIENT_URL=<fortio svc EXTERNAL_IP:port>

python ./runner/fortio.py $FORTIO_CLIENT_URL
```

### 6. Visualize results 

```
python ./runner/graph.py <PATH_TO_CSV> <METRIC> --mesh=linkerd 
```

#### Example: 

```
python ./runner/graph.py linkerd.csv p99 --mesh=linkerd 
```

![example-linkerd](example-linkerd-p99.png)
# Runner

This subdirectory contains the Python3 _module_ for automating topology
tests. The executable "main" for this is at "../run_tests.py".

## Pseudocode

```txt
read configuration
create cluster
add prometheus
for each topology:
  convert topology to Kubernetes YAML
  for each environment (none, istio, sidecars only, etc.):
    update Prometheus labels
    deploy environment
    deploy topology
    run load test
    delete topology
    delete environment
delete cluster
```

"""Module for automating topology testing.

The pseudo-code for the intended calls for this is:

```
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
"""

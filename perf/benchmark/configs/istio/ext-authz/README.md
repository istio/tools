# Ext-authz Benchmarking
This directory contains scripts and configurations for ext-authz benchmarking.
## Prerequisites
1. See [Istio Performance Benchmarking](https://github.com/istio/tools/tree/master/perf/benchmark) for environment setup.
2. Ext-authz setup: 
   - Deploy ext-authz in `twopods-istio` namespace
   - For more, see the [guide](https://istio.io/latest/docs/tasks/security/authorization/authz-custom/).

## Run the tests
To start the tests,
```
./run.sh
```
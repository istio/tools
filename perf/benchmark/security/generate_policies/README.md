# Istio Security Policy Generator

This directory contains information needed to create large scale security policies.

See the [Istio Security](https://istio.io/latest/docs/reference/config/security/) for more information about policies.

## Setup

To run generate_policies, run the following command:

```bash
go run generate_policies.go generate.go
```

This will by default create an Authorization Policy as follows and print it out to the stdout. This AuthorizationPolicy is specifically made to work with the environment that is created in the setup of [Istio Performance Benchmarking](https://github.com/istio/tools/tree/master/perf/benchmark)

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: test-1
  namespace: twopods-istio
spec:
 action: DENY
 rules:
 - from:
   - source:
       ipBlocks:
       - 0.0.0.0
```

generate_polices allows to customize the generated policies with command line flags:

```bash
Optional arguments:
  -h, --help
  -generate_policy string         List of key value pairs separated by commas.
                                  Supported options: namespace:string, action:DENY/ALLOW, policyType:AuthorizationPolicy, numPolicies:int, numPaths:int, numValues:int, numSourceIP:int, numNamespaces:int (default "numPolicies:1")
```

To create a large policy to an output .yaml file, run the following command:

```bash
go run generate_policies.go generate.go -generate_policy="numSourceIP:1000,numPaths:1000,numNamespaces:1000" > largePolicy.yaml
```

To apply largePolicy.yaml that was just created to istio use the following command.

```bash
kubectl apply -f largePolicy.yaml
```

## Example 1

```bash
go run generate_policies.go generate.go -generate_policy="numPolicies:10,numSourceIP:10,numPaths:2"
```

- This creates 10 AuthorizationPolicies which each contains 10 sourceIP's sources, and 2 paths operations

## Example 2

```bash
go run generate_policies.go generate.go -generate_policy="numSourceIP:100,numPaths:100,numNamespaces:100"
```

- This creates 1 AuthorizationPolicy which contains 100 sourceIP's sources, 100 paths operations, and 100 namespaces sources

## Cleanup

To remove the policies applied navigate to the generate_policies folder and run the following command (updating "largePolicy.yaml" if applied a different .yaml file): 

```bash
kubectl delete -f largePolicy.yaml
```

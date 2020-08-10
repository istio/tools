# Istio Security Policy Generator

This directory contains information needed to create large scale security policies.

See the [Istio Security](https://istio.io/latest/docs/reference/config/security/) for more information about policies.

The default values of the policies are specifically made to work with the environment that is created in the setup of [Istio Performance Benchmarking](https://github.com/istio/tools/tree/master/perf/benchmark)

## AuthorizationPolicy

To create an AuthorizationPolicy policy one must explicity pass in the following flag -generate_policy="AuthorizationPolicy:1".

To create an AuthorizationPolicy, run the following command:

```bash
go run generate_policies.go generate.go -generate_policy="AuthorizationPolicy:1,numSourceIP:1"
```

This will create an Authorization Policy as follows and print it out to the stdout. The number after "AuthorizationPolicy" is used to determine how many AuthorizationPolicies will be produced to stdout.

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: test-AuthorizationPolicy-1
  namespace: twopods-istio
spec:
 action: DENY
 rules:
 - from:
   - source:
       ipBlocks:
       - 0.0.0.0
```

The flags which can be used to create custom AuthorizationPolicies are as follows:

```bash
  AuthorizationPolicies:int
  namespace:string
  action:DENY/ALLOW
  numPolicies:int
  numPaths:int
  numValues:int
  numSourceIP:int
  numNamespaces:int
  numPrincipals:int
```

For more information see [AuthorizationPolicy Reference](https://istio.io/latest/docs/reference/config/security/authorization-policy/).

## PeerAuthentication

To create a PeerAuthentication policy one must explicity pass in the following flag -generate_policy="PeerAuthentication:1".

To create a PeerAuthentication, run the following command:

```bash
go run generate_policies.go generate.go -generate_policy="PeerAuthentication:1"
```

This will create a PeerAuthentication policy as follows since mtlsMode has a default value of STRICT and print it out to the stdout. The number after "PeerAuthentication" is used to determine how many PeerAuthentication will be produced to stdout.

```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: test-PeerAuthentication-1
  namespace: twopods-istio
spec:
 mtls:
   mode: STRICT
```

The flags which can be used to create custom PeerAuthentication are as follows:

```bash
  PeerAuthentication:int
  namespace:string
  mtlsMode:STRICT/DISABLE
```

For more information see [PeerAuthentication Reference](https://istio.io/latest/docs/reference/config/security/peer_authentication/).

## Examples

generate_policies.go also allows a user to create mutliple kinds of policies in one command.
To generate 1 AuthorizationPolicy with a principals rule and 1 PeerAuthorization policy with STRICT mtlsMode, run the following command:

```bash
 go run generate_policies.go generate.go -generate_policy="AuthorizationPolicy:1,numPrincipals:1,PeerAuthentication:1"
```

Which outputs the following yaml:

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: test-AuthorizationPolicy-1
  namespace: twopods-istio
spec:
 action: DENY
 rules:
 - from:
   - source:
       principals:
       - cluster.local/ns/twopods-istio/sa/Invalid-0
---
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: test-PeerAuthentication-1
  namespace: twopods-istio
spec:
 mtls:
   mode: STRICT
```

### Output to a yaml file

To create a large AuthorizationPolicy to an output .yaml file, run the following command:

```bash
go run generate_policies.go generate.go -generate_policy="AuthorizationPolicy:1,numSourceIP:1000,numPaths:1000,numNamespaces:1000" > largePolicy.yaml
```

### Apply the yaml file

To apply largePolicy.yaml that was just created to istio use the following command.

```bash
kubectl apply -f largePolicy.yaml
```

## Example 1

```bash
go run generate_policies.go generate.go -generate_policy="AuthorizationPolicy:1,numPolicies:10,numSourceIP:10,numPaths:2"
```

- This creates 10 AuthorizationPolicies which each contains 10 sourceIP's sources, and 2 paths operations

## Example 2

```bash
go run generate_policies.go generate.go -generate_policy="AuthorizationPolicy:1,numSourceIP:100,numPaths:100,numNamespaces:100"
```

- This creates 1 AuthorizationPolicy which contains 100 sourceIP's sources, 100 paths operations, and 100 namespaces sources

## Example 3

```bash
go run generate_policies.go generate.go -generate_policy="PeerAuthentication:1,mtlsMode:DISABLE"
```

- This creates 1 PeerAuthentication policy which has the mtls mode set to DISABLE

## Cleanup

To remove the policies applied navigate to the generate_policies folder and run the following command (updating "largePolicy.yaml" if applied a different .yaml file):

```bash
kubectl delete -f largePolicy.yaml
```

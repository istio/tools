
# Istio Security Policy Generator

This directory contains information needed to create large scale Security policies.

See the [Istio Security](https://istio.io/latest/docs/reference/config/security/) for more information about policies

## Setup

Build the project in the generate_policies folder

```bash
go build *.go
```
to run generate_polices run
```bash
./generate_polices
```
This will by default create an Authorization Policy as follows and print it out into the terminal. This AuthorizationPolicy is specifically made to work with the environment that is created in the setup of [Istio Performance Benchmarking](https://github.com/istio/tools/tree/master/perf/benchmark)

```bash
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: test_1
  namespace: twopods-istio
spec:
 action: DENY
 rules:
 - when:
   - key: request.headers[x-token]
     notValues:
     - admin
     - guest
 - to:
   - operation:
       methods:
       - GET
       - HEAD
       paths:
       - /admin
 - from:
   - source:
       namespaces:
       - twopods-istio
```
 generate_polices does allow for multiple optional command line flags which will allow you to customize you policies, these include
```bash
Optional arguments:
  -h, --help 
  -action string         Type of action (default "DENY")
  -namespace string      Current namespace (default "twopods-istio")
  -numPolicies int       Number of policies wanted (default 1)
  -policyType string     The type of security policy (default "AuthorizationPolicy")
  -to int                Number of To operations wanted (default 1)
  -when int              Number of when condition wanted (default 1)
  -from int              Number of From sources wanted (default 1)

```
To apply a rule created to istio one can apply the following command replacing the "POLICY" with the policy created.
```bash
kubectl apply -f - <<EOF
"POLICY"
EOF

```

## Example 1

```bash
 ./generate_polices -numPolicies=10 -to=10 -when=2
```

 - This creates 10 AuthorizationPolicies which each contain 10 "To" operations, 2 "When" conditions, and 1 "From" sources
## Example 2
```bash
 ./generate_polices -to=100 -when=100 -from=100
```
 - This creates 1 AuthorizationPolicy which each contain 100 "To" operations, 100 "When" conditions, and 100 "From" sources

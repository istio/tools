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

generate_polices allows to customize the generated policies with command line flags:

```bash
Optional arguments:
  -h, --help
  -generate_policy string         List of key value pairs separated by commas.
                                  Supported options: namespace:string, action:DENY/ALLOW, policyType:AuthorizationPolicy, numPolicies:int, when:int, from:int, to:int  (default "numPolicies:1")
```

To create a large policy to an output .yaml file, run the following command:

```bash
go run generate_policies.go generate.go -generate_policy="to:1000,when:1000,from:1000" > largePolicy.yaml
```

To apply largePolicy.yaml that was just created to istio use the following command.

```bash
kubectl apply -f largePolicy.yaml
```

## Example 1

```bash
go run generate_policies.go generate.go -generate_policy="numPolicies:10,to:10,when:2,from:1"
```

- This creates 10 AuthorizationPolicies which each contains 10 "To" operations, 2 "When" conditions, and 1 "From" sources

## Example 2

```bash
go run generate_policies.go generate.go -generate_policy="to:100,when:100,from:100"
```

- This creates 1 AuthorizationPolicy which each contains 100 "To" operations, 100 "When" conditions, and 100 "From" sources

## Run with runner.py

To measure the performance of having certain policies that have been applied, one can use [Istio Performance Benchmarking](https://github.com/istio/tools/tree/master/perf/benchmark) in particular the Run performance tests section with an extra flag (--security_policy) which will:
 1.  Generate policies depending on what was passed in as the arguments for the flag --security_policy and save the policies in a file called generated_policy.yaml within the generate_policies folder
 2. Apply those policies
 3. Run the performance test

To create specific policies the value which the security_policy flag will be assigned to, will be in the same format as in the examples above for -generate_policy:

```bash
"to:100,when:100,from:100"
```

#### Cleanup

To remove the policies applied navigate to the generate_policies folder and run the following command:

```bash
kubectl delete -f generated_policy.yaml
```

## Runner.py Example 1

```bash
python3 runner.py --conn 64 --qps 1000 --duration 240 --baseline --load_gen_type=fortio --telemetry_mode=v2-nullvm --security_policy="numPolicies:1,from:100"
```

- This creates 2 AuthorizationPolicies which each contain 100 from rules, applies those policies, then runs the performance test.

The example output should start with:

```bash
authorizationpolicy.security.istio.io/test-1 created
-------------- Running in baseline mode --------------
```

## Runner.py Example 2

```bash
python3 runner.py --conn 64 --qps 1000 --duration 240 --baseline --load_gen_type=fortio --telemetry_mode=v2-nullvm --security_policy="numPolicies:10,from:1"
```

- This creates 10 AuthorizationPolicies which each contain 1 from rules, applies those policies. Then runs the performance test

The example output should start with:

```bash
authorizationpolicy.security.istio.io/test-1 created
authorizationpolicy.security.istio.io/test-2 created
authorizationpolicy.security.istio.io/test-3 created
authorizationpolicy.security.istio.io/test-4 created
authorizationpolicy.security.istio.io/test-5 created
authorizationpolicy.security.istio.io/test-6 created
authorizationpolicy.security.istio.io/test-7 created
authorizationpolicy.security.istio.io/test-8 created
authorizationpolicy.security.istio.io/test-9 created
authorizationpolicy.security.istio.io/test-10 created
-------------- Running in baseline mode --------------
```

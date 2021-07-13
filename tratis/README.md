# Tratis

Tratis (**tr**acing **to** **is**otope) converts distributed tracing traces
into isotope configuration files so that developers and users can
model realistic applications.

Tratis accepts distributed trace data produced by distributed tracing tools
(Jaeger) and is able to capture

+ Application Call Order (Dependency Graph, Call Order) AS a n-ary tree.
+ Request Processing Times AS known distributions.

## Setup

+ Run `pip3 install -r runner/requirements.txt`
+ Setup a microservice
+ Setup Istio
+ Setup a distributed tracer i.e. Jaeger
+ Run `go run service/main.go <Traces.json> <Output.json>`

## File Information

### Trace.json

This file will contain the raw trace data captured by the distributed tracer.

### Output.json

This file will contain the tratis output:

For each set of traces (trace that have the same dependency graph):

+ An annotated n-ary tree
+ Time Information
    + Distribution
    + Mean + STD
+ Request Size Information
    + Distribution
    + Mean + STD
+ Respones Size Information
    + Distribution
    + Mean + STD
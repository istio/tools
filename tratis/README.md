# Tratis

Tratis (**tr**acing **to** **is**otope) converts distributed tracing traces
into isotope configuration files so that developers and users can
model realistic applications.

Tratis accepts distributed trace data produced by distributed tracing tools
(Jaeger, Zipkin) and is able to capture

1. Application Call Order (Dependency Graph, Call Order) AS a n-ary tree.
2. Request Processing Times AS known distributions.

Currently Tratis only supports applications which don't have a dynamic
call graph i.e. the call graph remains the same for each request.
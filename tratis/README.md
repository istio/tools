# Tratis

Tratis (**tr**acing **to** **is**otope) converts distributed tracing traces
into isotope configuration files so that developers and users can
model realistic applications.

Tratis accepts distributed trace data produced by distributed tracing tools
(Jaeger) and is able to capture

+ Application Call Order (Dependency Graph, Call Order) AS a n-ary tree.
+ Request Processing Times AS known distributions.
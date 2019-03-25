# External Traffic

This tests external traffic without using `ServiceEntry`s.

Instead, `global.outboundTrafficPolicy.mode=ALLOW_ANY` is set and a Sidecar resource is created.

This tests that we can send traffic from a deployment to the ingress gateway (simulating external traffic), back to another deployment in the cluster.

Additionally, we create an unrelated ServiceEntry, as well as a Sidecar resource that does not import this entry, to ensure it does not interfere.
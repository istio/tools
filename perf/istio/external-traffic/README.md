# External Traffic

This tests external traffic without using `ServiceEntry`s.

Instead, `global.outboundTrafficPolicy.mode=ALLOW_ANY` is set and a Sidecar resource is created.

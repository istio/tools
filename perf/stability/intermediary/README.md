# Intermediary Test

Intermediary middle proxy test.
-------------------------------

Istio can be used to forward traffic meant for a service through an intermediary.
Typical use case is content filtering via mod-security. We set the default route for
a destination to the intermediary, and the intermediary must forward it to its destination.
This setup is similar to egress proxy setup, except we are injecting a sidecar along with
the content filtering proxy.


This test mocks an intermediate content filtering proxy.
It uses envoy to implement the mock.

1. Client - fortio client issues http requests.
2. Intermediary - This proxy rejects requests if the post body contains a specific string.
           Otherwise it modifies the request path and forwards the request.

           The intermediate proxy *must not* change the host header.
           The sidecar next to the intermediate proxy will route the request to the correct destination
           based on the host header / cluster VIP.

           The sidecar injected in the intermediary pod handles TLS and routing, freeing the intermediary to behave
           like a bump-on-the-wire proxy.
3. Server - httpbin server.

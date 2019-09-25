# Intermediary Test

## Intermediary middle proxy test

Istio can be used to forward traffic meant for a service through an intermediary.
Typical use case is content filtering via mod-security. We set the default route for
a destination to the intermediary, and the intermediary must forward it to its destination.
This setup is similar to egress proxy setup, except we are injecting a sidecar along with
the content filtering proxy.

This test mocks an intermediate content filtering proxy.
It uses envoy to implement the mock.

### Client

A Fortio client that issues http requests.

### Intermediary

The intermediary proxy rejects requests if the post body contains a specific string.
Otherwise it modifies the request uri and forwards the request.

An intermediate proxy *must not* change the host header. The sidecar associated with the intermediate
proxy will route the request to the correct destination based on the host header and cluster VIP.
The sidecar also handles mTLS, routing, and load balancing freeing the intermediary to behave like a bump-on-the-wire proxy.

### Server

An httpbin server.

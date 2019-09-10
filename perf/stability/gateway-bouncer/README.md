# Gateway Bouncer

This scenario spins up a minimal 'isolated' setup that can be used for testing
or the IngressGateway in ways that are either destructive or require setup
that is not compatible with the default gateways under `istio-system`.

One the setup is up and running the following capabilities are exercised:

1. External traffic is simulated by Fortio and will go via external load
balancer through the IngressGateway (a new LoadBalancer will be created
automatically by the install script and its external IP add`ress will be used).
1. A rolling restart of the IngressGateway will be periodically triggered while
the corresponding Pilot is down (and eventually brought up), so the readiness
checks of the IngressGateway can be verified. IngressGateway should not become
ready until the Pilot is back up and a configuration has been received.
1. Rolling restart of the gateway should be transparent for the clients. Fortio
Client is configured to crash upon connectivity errors. 'Connection refused'
will result in a restart of Fortio Client pods. Prometheus metrics on pod
status can be used to verify whether rolling redeploy of the IngressGateway
results in Fortio Client restarts.
1. The IngressGateway routes are configured using standard K8S Ingress with a
custom Ingress class label to make sure multiple Ingress configurations can
co-exist in a single cluster.

Use the following command to (re)install the scenario into a new/existing
namespace (set `DRY_RUN=1` for a dry run):

```bash
NAMESPACE=gateway-bouncer ./setup.sh
```

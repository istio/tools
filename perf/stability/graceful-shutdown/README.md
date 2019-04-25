# Graceful Shutdown Test

This test ensures that proxies will be shutdown gracefully.

This is measured by sending many long lasting requests.

When the server is redeployed, traffic should gracefully transition to the new deployment - connections should not be dropped.

It is recommended to also run the `istio-upgrader` test or the `istio-chaos`, to ensure connections are able to handle Istio control plane upgrades and downtime.

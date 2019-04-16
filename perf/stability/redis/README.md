# Redis

This test installs an instance of Redis using the [stable/redis](https://github.com/helm/charts/tree/master/stable/redis) Helm chart.

The Redis install is generated using `helm template stable/redis --set password=istio --name redis`

Additionally, a simple redis client is created that repeatedly writes to the master instance, then tries to read that value from the master and slave instance.

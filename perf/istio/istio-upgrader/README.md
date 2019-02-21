# Istio Upgrader Test

This test periodically triggers an upgrade of Istio.

ImagePullPolicy should be set to Always to get the most impact from this test, as this will pull down new updates. This is set already if Istio is installed with the perf test setup.
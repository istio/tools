# Istio Redeployment Test

This test periodically triggers a redeployment of Istio.

ImagePullPolicy should be set to Always to get the most impact from this test, as this will pull down new updates if you are using an image like `latest-daily`. This is set already if Istio is installed with the perf test setup.
# Kafka

This test installs an instance of Kafka using the [incubator/kafka](https://github.com/helm/charts/tree/master/incubator/kafka) Helm chart.

The charts are based on the `incubator/kafka` charts, with some changes:
* An `istio-headless.yaml` file is added, to support using the headless Kafka services. Istio will not generate an entry for each pod in the headless service by default, so we need to do this manually.
* Kafka and Zookeeper expose the same port in a headless and standard service, which can cause conflicts. To avoid this, we use a different port on the standard (public) service.
* Istio Sidecar is disabled for Zookeeper. It should be possible to get Zookeeper working, but for now this example showcases strictly Kafka with Istio.

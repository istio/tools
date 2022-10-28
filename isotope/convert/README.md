# Converter

This subdirectory contains the Go command which converts topology YAML to
various formats.

The root main.go outputs a [Cobra](https://github.com/spf13/cobra) CLI for
controlling the behavior of the program.

## Conversion Outputs

- __Kubernetes__ (`go run main.go kubernetes <topology_path> ...`):
  Generates services and deployments for all topology services and the
  [Fortio](https://github.com/istio/fortio) client to load test against them.

To generate the output for isotope mock services:

```shell
go run main.go kubernetes <topology_path> --service-image <isotope-service-image>  > output.yaml
```

To see all the options:

```shell
go run main.go kubernetes  --help
```

### Using the examples

To generate the output for 2 services in a different namespace:

```shell
go run main.go kubernetes ../example-topologies/chain-2-services-different-namespaces.yaml \
    --service-image <isotope-service-image>  > output.yaml
```

To generate the output for specific services belonging to a given cluster name:

```shell
# For cluster1
go run main.go kubernetes ../example-topologies/chain-2-services-different-cluster.yaml \
    --service-image <isotope-service-image> --cluster cluster1  > cluster1.yaml

# For cluster2
go run main.go kubernetes ../example-topologies/chain-2-services-different-cluster.yaml \
    --service-image <isotope-service-image> --cluster cluster2  > cluster2.yaml
```

### Using the container

```shell
docker run -it \
    -v $(pwd)/../example-topologies/chain-2-services-different-namespaces.yaml:/etc/config/service-graph.yaml  \
    <isotope-convert-image>  kubernetes --service-image <isotope-service-image>  > output.yaml
```

## Deploy

You can build and deploy the image by your own, or to build and push the image
with [ko](https://github.com/ko-build/ko)
and a [ephemeral registry](https://www.civo.com/learn/ttl-sh-your-anonymous-and-ephemeral-docker-image-registry)
in one command like follows:

```shell
export KO_DOCKER_REPO=ttl.sh/<my-prefix>-isotope-convert
ko build --bare -t 1h .
```

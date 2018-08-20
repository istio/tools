# Converter

This subdirectory contains the Go command which converts topology YAML to
various formats.

The root main.go outputs a [Cobra](https://github.com/spf13/cobra) CLI for
controlling the behavior of the program.

## Conversion Outputs

- __Graphviz__ (`go run main.go graphviz <topology_path> <output>`):
  Generates [Graphviz](https://www.graphviz.org) [DOT
  language](https://www.graphviz.org/doc/info/lang.html)
- __Kubernetes__ (`go run main.go kubernetes <topology_path> ...`):
  Generates services and deployments for all topology services and the
  [Fortio](https://github.com/istio/fortio) client to load test against them.

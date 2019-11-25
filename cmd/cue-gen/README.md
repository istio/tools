# cue-gen

`cue-gen` is a tool that generates OpenAPI schema and Kubernetes `CustomResourceDefinition`(CRD) configurations. It relies on [`cuelang`](https://cuelang.org/) packages to translate
Protobuf definitions to OpenAPI schemas, specifically
[structural schemas](https://kubernetes.io/docs/tasks/access-kubernetes-api/custom-resources/custom-resource-definitions/#specifying-a-structural-schema) in CRDs.

## Usage

`cue-gen` can be used to generate OpenAPI schemas for each package respectively, OpenAPI schemas of all packages in one file, and CRDs in a file.

### Generate OpenAPI schemas

To generate the OpenAPI schemas, configurations need to be specified in a JSON or YAML file. Use `cue-gen -help` to find out the configuration options.

#### Generate for each package respectively

`directories` need to be specified in the configuration file for packages that need OpenAPI schemas. Use `cue-gen -f={PATH_TO_CONFIG_FILE} -paths={PATH_TO_PROTO_IMPORTS}`
to get the OpenAPI schema file(s) in each package.

#### Generate for all packages in one file

In addition to `directories` field in the configuration file, the `all` field needs to be specified. Use `cue-gen -all -f={PATH_TO_CONFIG_FILE} -paths={PATH_TO_PROTO_IMPORTS}`
to get the OpenAPI schemas in a single file.

### Generate CRDs

Configurations on how CRDs are generated are specified in the comments of the protos that map to the CRDs. For example, to generate CRD for `DestinationRule`, the following configuration
needs to be added to the comment of the `DestinationRule` proto.

``` protobuf
// <!-- crd generation tags
// +cue-gen:DestinationRule:groupName:networking.istio.io
// +cue-gen:DestinationRule:version:v1alpha3
// +cue-gen:DestinationRule:storageVersion
// +cue-gen:DestinationRule:annotations:helm.sh/resource-policy=keep
// +cue-gen:DestinationRule:labels:app=istio-pilot,chart=istio,heritage=Tiller,release=istio
// +cue-gen:DestinationRule:subresource:status
// +cue-gen:DestinationRule:scope:Namespaced
// +cue-gen:DestinationRule:resource:categories=istio-io,networking-istio-io,shortNames=dr
// +cue-gen:DestinationRule:printerColumn:name=Host,type=string,JSONPath=.spec.host,description="The name of a service from the service registry"
// +cue-gen:DestinationRule:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp
// representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations.
// Clients may not set this value. It is represented in RFC3339 form and is in UTC.
// Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
// -->
```

Use `cue-gen -crd -f={PATH_TO_CONFIG_FILE} -paths={PATH_TO_PROTO_IMPORTS}` to get the CRD file at `{PWD}/kubernetes/customresourcedefinitions.gen.yaml`.

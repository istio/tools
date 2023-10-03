# `protoc-gen-crd`

`protoc-gen-crd` is a plugin for protobufs to compile to Kubernetes CRD (OpenAPI) schemas.

This command is a fork of <https://github.com/solo-io/protoc-gen-openapi>, which originated the Protobuf -> OpenAPI logic that forms the basis for this command.

Along with general changes to support CRDs (and removal of pieces not needed for CRDs), this fork is highly Istio opinionated, hence the fork.
In part, this maintains compatibility with the older CRD generation mechanism, `cue-gen`.

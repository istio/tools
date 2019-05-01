# mesh-wide RPS load testing

This directory contains tools to generate a user-defined count of fortio services and
user defined ccount of fortio clients that communicate internally in the mesh to measure
mesh wide RPS via grafana.  Each service connects to a unique fortio deployment, and each
client is a unique fortio deployment.

## Introduction

These charts generate manifests of arbitrary size.  Take care not to exceed the system
capacity.  A single node kubeadm cluster can handle 110 pods maximum, but this can be
overridden.

## Setup

Install Istio with the values.yaml of your choice.  This PR was tested on a dual Xeon
32 core + 128 gb bare metal machine with Istio 1.1.3 defaults with the `istio-values.yaml`
in this directory as overrides with the `helm -f` command.  the value `20` was tested for
`NUM`

To setup, run `./generate_manifests.sh NUM`, where `NUM` is the clients and servers count
to generate.  For each increment of `NUM`, 4 pods are created and 1 service is created.

1. Label the default namespace to enable automamtic injection
1. run `kubectl apply -f server-manifest.yaml`
1. run `kubectl apply -f client-manifest.yaml`

Services and deployments will be created in the default namespace.

## Cleanup

1. run `kubectl delete -f client-manifest.yaml`
1. run `kubectl delete -f server-manifest.yaml`

## More details

This is meant to run with the default namespace labeled.  Running without the default
namespace labeled will result in unexpected behavior.  Additionally this is meant to run
with TLS enabled.

Unlike nearly every other testing tool in this repository, this does not test ingress,
but instead tests internal mesh communication.


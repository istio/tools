#!/bin/bash

set -xe

helm template helm/client --set namespace=$1 --set entries=$2 > client-manifest.yaml
helm template helm/server --set namespace=$1 --set entries=$2 > server-manifest.yaml

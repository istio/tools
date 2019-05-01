#!/bin/bash

set -xe

helm template helm/client --set entries=$1 > client-manifest.yaml
helm template helm/server --set entries=$1 > server-manifest.yaml

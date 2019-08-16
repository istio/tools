#!/usr/bin/env bash

HUB=gcr.io/istio-testing
VERSION=$(date +%Y-%m-%dT%H-%M-%S)

docker build --no-cache -t $HUB/build-tools:$VERSION .
docker push $HUB/build-tools:$VERSION

#!/usr/bin/env bash

HUB=gcr.io/istio-testing
VERSION=$(date +%Y-%m-%d)

docker build --no-cache -t $HUB/build-tools:$VERSION .
docker push $HUB/build-tools:$VERSION

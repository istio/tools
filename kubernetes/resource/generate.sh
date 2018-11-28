#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

pushd ${DIR}/../..

protoc --go_out=plugins=grpc:../.. -I=. kubernetes/resource/options.proto

popd
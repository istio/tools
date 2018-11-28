#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

pushd ${DIR}
go build
popd

pushd ${DIR}/..

rm -rf ${DIR}/testdata/out
mkdir ${DIR}/testdata/out

protoc --go_out=plugins=grpc:../.. -I=. protoc-gen-crds/testdata/basic.proto
protoc --plugin=protoc-gen-crds/protoc-gen-crds --proto_path=. --crds_out=protoc-gen-crds/testdata/out protoc-gen-crds/testdata/basic.proto


CODEGEN_PATH=${DIR}/../../../k8s.io/code-generator
pushd ${CODEGEN_PATH}

./generate-groups.sh "deepcopy,client,informer,lister" \
  istio.io/tools/protoc-gen-crds/testdata/out \
  istio.io/tools/protoc-gen-crds/testdata/out \
  config.istio.io:v1alpha1 \
  --go-header-file ${DIR}/boilerplate.txt
popd


popd



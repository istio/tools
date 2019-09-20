
go_package = istio.io/tools/cmd/protoc-gen-deepcopy/test/generated
import_map = Mtypes.proto=$(go_package),Mexternal.proto=$(go_package)

all: clean test

clean:
	if [ -d "generated" ]; then rm -rf generated; fi

test: generate gobuild gotest

generate:
	if [ ! -d "generated" ]; then mkdir generated; fi
	protoc --gogofast_out=plugins=grpc:./generated --deepcopy_out=:./generated types.proto

gobuild:
	go build ./...

gotest:
	go test .


go_package = istio.io/tools/cmd/protoc-gen-deepcopy/test/generated

all: clean test

clean:
	if [ -d "generated" ]; then rm -rf generated; fi

test: generate gobuild gotest

generate:
	if [ ! -d "generated" ]; then mkdir generated; fi
	protoc --go_out=./generated --deepcopy_out=:./generated types.proto

gobuild:
	go build ./...

gotest:
	go test .

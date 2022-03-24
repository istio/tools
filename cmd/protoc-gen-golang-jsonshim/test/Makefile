
go_package = istio.io/tools/cmd/protoc-gen-golang-jsonshim/test/generated

all: clean test

clean:
	if [ -d "generated" ]; then rm -rf generated; fi

test: generate gobuild gotest

generate:
	if [ ! -d "generated" ]; then mkdir generated; fi
	protoc --go_out=./generated --golang-jsonshim_out=:./generated types.proto external.proto

gobuild:
	go build ./...

gotest:
	go test .


go_package = istio.io/tools/cmd/protoc-gen-golang-jsonshim/test/generated

all: clean test

clean:
	if [ -d "generated" ]; then rm -rf generated; fi

test: generate gobuild gotest

generate:
	if [ ! -d "generated" ]; then mkdir generated; fi
	protoc --go_out=. --go_opt=paths=source_relative \
		   --alias_out=. --alias_opt=paths=source_relative \
		   v1/*.proto

gobuild:
	go build ./...

gotest:
	go test .

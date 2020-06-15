
go_package = istio.io/tools/cmd/cue-gen/test/generated
import_map = Mtypes.proto=$(go_package),Mexternal.proto=$(go_package)

all: clean test

clean:
	if [ -d "generated" ]; then rm -rf generated; fi

test: generate gobuild gotest

generate:
	if [ ! -d "generated" ]; then mkdir generated; fi
	../cue-gen -f=istio.io/api/cue.yaml -crd=true -status=istio.meta.v1beta1.Status

gobuild:
	go build ./...

gotest:
	go test .


all: build run

build:
	go build -o protoc-gen-openapi

run:
	rm -fr out
	mkdir out
	protoc --plugin=./protoc-gen-openapi --openapi_out=mode=true,single_file=true,use_ref=true:out/. -Itestdata testdata/testpkg/test1.proto testdata/testpkg/test2.proto testdata/testpkg/test6.proto testdata/testpkg2/test3.proto

gotest:
	go test

clean:
	@rm -fr out protoc-gen-openapi

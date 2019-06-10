
all: build run

build:
	go build -o protoc-gen-openapi

run:
	rm -fr out
	mkdir out
	protoc --plugin=./protoc-gen-openapi --openapi_out=mode=true:out/. -Itest/testdata test/testdata/testpkg/test1.proto test/testdata/testpkg/test2.proto test/testdata/testpkg/test6.proto test/testdata/testpkg2/test3.proto

gotest:
	go test ./test

test: build gotest

clean:
	@rm -fr out protoc-gen-openapi

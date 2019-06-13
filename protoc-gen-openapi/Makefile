
all: build run

build:
	go build -o protoc-gen-openapi

run:
	rm -fr out
	mkdir out
	protoc --plugin=./protoc-gen-openapi --openapi_out=mode=true:out/. testdata/test1.proto testdata/test2.proto testdata/test3.proto

clean:
	@rm -fr out protoc-gen-openapi


all: build run

build:
	go build -o protoc-gen-docs *.go

run:
	rm -fr fm page fragment
	mkdir fm page fragment
	protoc --plugin=./protoc-gen-docs --docs_out=warnings=false,emit_yaml=true,mode=html_page:page/. testdata/test1.proto testdata/test2.proto testdata/test3.proto

more: all
	protoc --plugin=./protoc-gen-docs --docs_out=warnings=false,mode=html_fragment_with_front_matter:fm/. testdata/test1.proto testdata/test2.proto 	testdata/test3.proto
	protoc --plugin=./protoc-gen-docs --docs_out=warnings=false,mode=html_fragment:fragment/. testdata/test1.proto testdata/test2.proto testdata/test3.proto

clean:
	@rm -fr fm page fragment protoc-gen-docs

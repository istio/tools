
all: build run

build:
	go build -o protoc-gen-docs *.go

run:
	rm -fr fm page fragment pf sp
	mkdir fm page fragment pf sp
	protoc -Iprotos -I. --plugin=./protoc-gen-docs --docs_out=warnings=false,emit_yaml=true,mode=html_page:page/. testdata/test1.proto testdata/test2.proto testdata/test3.proto
	protoc -Iprotos -I. --plugin=./protoc-gen-docs --docs_out=warnings=false,mode=html_fragment_with_front_matter:fm/. testdata/test1.proto testdata/test2.proto testdata/test3.proto
	protoc -Iprotos -I. --plugin=./protoc-gen-docs --docs_out=warnings=false,mode=html_fragment:fragment/. testdata/test1.proto testdata/test2.proto testdata/test3.proto
	protoc -Iprotos -I. --plugin=./protoc-gen-docs --docs_out=warnings=true,per_file=true,mode=html_fragment_with_front_matter:pf/. testdata/test1.proto
	protoc -Iprotos -I. --plugin=./protoc-gen-docs --docs_out=warnings=true,dictionary=dictionaries/en-US,custom_word_list=dictionaries/custom.txt,mode=html_fragment_with_front_matter:sp/. testdata/test6.proto

clean:
	@rm -fr fm page fragment pf sp sp2 protoc-gen-docs

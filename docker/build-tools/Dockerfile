FROM golang:1.12 as build_tools

# Pinned versions of stuff we pull in
ENV PROTOBUF_VERSION=3.6.1
ENV GOGO_PROTOBUF_VERSION=v1.2.2
ENV GOLANG_PROTOBUF_VERSION=v1.3.1
ENV PROTOLOCK_VERSION=v0.14.0
ENV PROTOTOOL_VERSION=v1.8.0
ENV COUNTERFEITER_VERSION=v6.2.2
ENV GOIMPORTS_VERSION=379209517ffe
ENV GOLANGCI_LINT_VERSION=v1.16.0

# Update distro and install dependencies
RUN apt-get update
RUN apt-get install -y build-essential unzip autoconf libtool zlibc zlib1g-dev libc-ares-dev libssl-dev upx

# Prepare common output directories
RUN mkdir -p /usr/bin
RUN mkdir -p /usr/include

# Install protoc
WORKDIR /tmp
ADD https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip /tmp/
RUN unzip /tmp/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip
RUN cp -aR /tmp/bin/protoc /usr/bin

# Install necessary protoc includes
RUN mkdir -p /usr/include/protobuf/google/protobuf
RUN mkdir -p /usr/include/protobuf/google/rpc
RUN mkdir -p /usr/include/protobuf/gogoproto
RUN for f in any duration descriptor empty struct timestamp wrappers; do \
            curl -L -o /usr/include/protobuf/google/protobuf/${f}.proto https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/${f}.proto; \
        done
RUN for f in code error_details status http; do \
            curl -L -o /usr/include/protobuf/google/rpc/${f}.proto https://raw.githubusercontent.com/istio/gogo-genproto/master/googleapis/google/rpc/${f}.proto; \
        done
RUN curl -L -o /usr/include/protobuf/gogoproto/gogo.proto https://raw.githubusercontent.com/gogo/protobuf/master/gogoproto/gogo.proto

# Install external dependencies
RUN GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@${GOLANG_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/istio/gogo-protobuf/protoc-gen-gogofast@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/istio/gogo-protobuf/protoc-gen-gogofaster@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/istio/gogo-protobuf/protoc-gen-gogoslick@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/uber/prototool/cmd/prototool@${PROTOTOOL_VERSION}
RUN GO111MODULE=on go get github.com/nilslice/protolock/cmd/protolock@${PROTOLOCK_VERSION}
RUN GO111MODULE=on go get github.com/maxbrunsfeld/counterfeiter/v6@${COUNTERFEITER_VERSION}
RUN GO111MODULE=on go get golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}
RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
RUN go get github.com/jteeuwen/go-bindata/go-bindata

# Install latest versions of custom Istio tools
RUN GO111MODULE=on go get istio.io/tools/protoc-gen-docs@latest
RUN GO111MODULE=on go get istio.io/tools/cmd/annotations_prep@latest

# Save some bytes in the container
RUN upx --lzma /go/bin/*

# Prewarm the golang module cache
RUN rm -fr /go/pkg/mod/*
RUN mkdir -p /home
WORKDIR /home
COPY go.mod /home
RUN go mod download

# Cleanup
WORKDIR /
RUN apt-get purge -y unzip autoconf libtool zlibc zlib1g-dev libc-ares-dev libssl-dev upx
RUN apt-get autoremove -y
RUN apt-get clean

# TODO: Remove this block if the FROM distroless below is made to work.
RUN rm -fr /root/.cache/go-build
RUN rm -fr /home
RUN rm -fr /tmp/*

# TODO: sadly this doesn't work as some processes launched in the container won't start. For example, make fails with the super useful:
#
#   make: *** [gen] Error 139
#
# There must be some additional dependency or variable that needs to be copied over
#
# Render the final image with minimal stuff, which saves over 300M compared to the fuller container image up to this point.
#FROM gcr.io/distroless/base:latest
#
#COPY --from=build_tools /usr/ /usr
#COPY --from=build_tools /go/ /go
#COPY --from=build_tools /bin/ /bin
#COPY --from=build_tools /lib/ /lib
#
#ENV GOPATH /go
#ENV PATH $GOPATH/bin:$PATH

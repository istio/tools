FROM golang:1.12 as build_tools_context

# Pinned versions of stuff we pull in
ENV PROTOBUF_VERSION=3.6.1
ENV GOGO_PROTOBUF_VERSION=28a6bbf47e48e0b2220b2a244750b660c83d4942
ENV GOLANG_PROTOBUF_VERSION=v1.3.1
ENV PROTOLOCK_VERSION=v0.14.0
ENV PROTOTOOL_VERSION=v1.8.0
ENV COUNTERFEITER_VERSION=v6.2.2
ENV GOIMPORTS_VERSION=379209517ffe
ENV GOLANGCI_LINT_VERSION=v1.16.0

ENV OUTDIR=/out

# Update distro and install dependencies
# TODO we probably don't need all this stuff...
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    unzip \
    upx \
 && rm -rf /var/lib/apt/lists/*

# Prepare common output directories
RUN mkdir -p /usr/bin /usr/include

# Install protoc
WORKDIR /tmp
ADD https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip /tmp/
RUN unzip /tmp/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip

# Install necessary protoc includes
RUN mkdir -p ${OUTDIR}/usr/include/protobuf/google/protobuf
RUN mkdir -p ${OUTDIR}/usr/include/protobuf/google/rpc
RUN mkdir -p ${OUTDIR}/usr/include/protobuf/gogoproto

RUN for f in any duration descriptor empty struct timestamp wrappers; do \
            curl -L -o ${OUTDIR}/usr/include/protobuf/google/protobuf/${f}.proto https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/${f}.proto; \
        done
RUN for f in code error_details status http; do \
            curl -L -o ${OUTDIR}/usr/include/protobuf/google/rpc/${f}.proto https://raw.githubusercontent.com/istio/gogo-genproto/master/googleapis/google/rpc/${f}.proto; \
        done
RUN curl -L -o ${OUTDIR}/usr/include/protobuf/gogoproto/gogo.proto https://raw.githubusercontent.com/gogo/protobuf/master/gogoproto/gogo.proto

# Install external dependencies
RUN GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@${GOLANG_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogofast@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogofaster@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogoslick@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/uber/prototool/cmd/prototool@${PROTOTOOL_VERSION}
RUN GO111MODULE=on go get github.com/nilslice/protolock/cmd/protolock@${PROTOLOCK_VERSION}
RUN GO111MODULE=on go get github.com/maxbrunsfeld/counterfeiter/v6@${COUNTERFEITER_VERSION}
RUN GO111MODULE=on go get golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}
RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
RUN go get github.com/jteeuwen/go-bindata/go-bindata

# Install latest versions of custom Istio tools
RUN go get istio.io/tools/protoc-gen-docs
RUN go get istio.io/tools/cmd/annotations_prep

# Put the stuff we need in its final output location
RUN mkdir -p ${OUTDIR}/go/bin
RUN mkdir -p ${OUTDIR}/usr/include

RUN cp -aR /tmp/bin/protoc ${OUTDIR}/go/bin
RUN cp -aR /tmp/include/* ${OUTDIR}/usr/include
RUN cp -aR /go/bin/* ${OUTDIR}/go/bin

# Use inplace decompression on the go toolchain binaries
RUN upx --lzma ${OUTDIR}/go/bin/*

# Create a golang toolchain context
FROM golang:1.12 as golang_context

# Build a base os context
FROM ubuntu:xenial as base_os_context

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    make \
    ca-certificates \
    git \
    wget \
    curl \
 && rm -rf /var/lib/apt/lists/*

# Setup to run the Go compiler with host-side build cache, test cache, and module cache
FROM scratch

ENV GO111MODULE=on
ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV PATH /go/bin:$PATH

COPY --from=base_os_context / /
COPY --from=build_tools_context /out/ /
COPY --from=golang_context /usr/local/go/ /go/

WORKDIR /

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
ENV PROTOC_GEN_VALIDATE_VERSION=d6164de4910977d3c3c8dbd9299b5064ea9e7c2b
ENV GO_BINDATA_VERSION=v3.0.8-0.20180305030458-6025e8de665b
ENV PROTOC_GEN_GRPC_GATEWAY_VERSION=v1.8.1
ENV PROTOC_GEN_SWAGGER_VERSION=v1.8.1

ENV OUTDIR=/out

# Update distro and install dependencies
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    unzip \
    upx

# Install protoc
WORKDIR /tmp
ADD https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip /tmp/
RUN unzip /tmp/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip

# Install necessary protoc includes
RUN mkdir -p ${OUTDIR}/usr/include/protobuf/google/protobuf
RUN mkdir -p ${OUTDIR}/usr/include/protobuf/google/rpc
RUN mkdir -p ${OUTDIR}/usr/include/protobuf/google/api
RUN mkdir -p ${OUTDIR}/usr/include/protobuf/gogoproto
RUN mkdir -p ${OUTDIR}/usr/include/protobuf/validate

RUN for f in any duration descriptor empty struct timestamp wrappers; do \
            curl -L -o ${OUTDIR}/usr/include/protobuf/google/protobuf/${f}.proto https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/${f}.proto; \
        done
RUN for f in code error_details status; do \
            curl -L -o ${OUTDIR}/usr/include/protobuf/google/rpc/${f}.proto https://raw.githubusercontent.com/istio/gogo-genproto/master/googleapis/google/rpc/${f}.proto; \
        done
RUN for f in http annotations; do \
            curl -L -o ${OUTDIR}/usr/include/protobuf/google/api/${f}.proto https://raw.githubusercontent.com/istio/gogo-genproto/master/googleapis/google/api/${f}.proto; \
        done
RUN curl -L -o ${OUTDIR}/usr/include/protobuf/validate/validate.proto https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/${PROTOC_GEN_VALIDATE_VERSION}/validate/validate.proto
RUN curl -L -o ${OUTDIR}/usr/include/protobuf/gogoproto/gogo.proto https://raw.githubusercontent.com/gogo/protobuf/master/gogoproto/gogo.proto

# Install a bunch of Go tools
RUN GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@${GOLANG_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gofast@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogofast@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogofaster@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogoslick@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/uber/prototool/cmd/prototool@${PROTOTOOL_VERSION}
RUN GO111MODULE=on go get github.com/nilslice/protolock/cmd/protolock@${PROTOLOCK_VERSION}
RUN GO111MODULE=on go get github.com/maxbrunsfeld/counterfeiter/v6@${COUNTERFEITER_VERSION}
RUN GO111MODULE=on go get golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}
RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
RUN GO111MODULE=on go get github.com/jteeuwen/go-bindata/go-bindata@${GO_BINDATA_VERSION}
RUN GO111MODULE=on go get github.com/envoyproxy/protoc-gen-validate@${PROTOC_GEN_VALIDATE_VERSION}
RUN GO111MODULE=on go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@${PROTOC_GEN_GRPC_GATEWAY_VERSION}
RUN GO111MODULE=on go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@${PROTOC_GEN_SWAGGER_VERSION}

# Install latest version of Istio-owned tools
RUN GO111MODULE=on go get istio.io/tools/protoc-gen-docs
RUN GO111MODULE=on go get istio.io/tools/cmd/annotations_prep
RUN GO111MODULE=on go get istio.io/tools/openapi/cue
RUN GO111MODULE=on go get istio.io/tools/cmd/vfsgen

# Put the stuff we need in its final output location
RUN mkdir -p ${OUTDIR}/go/bin
RUN mkdir -p ${OUTDIR}/usr/bin
RUN mkdir -p ${OUTDIR}/usr/include

RUN cp -aR /tmp/bin/protoc ${OUTDIR}/usr/bin
RUN cp -aR /tmp/include/* ${OUTDIR}/usr/include
RUN cp -aR /go/bin/* ${OUTDIR}/go/bin

# Use inplace decompression on the go binaries
RUN upx --lzma ${OUTDIR}/go/bin/*

# Create a golang toolchain context
FROM golang:1.12 as golang_context

# Cleanup stuff we don't need in the final image
RUN rm -fr /usr/local/go/doc
RUN rm -fr /usr/local/go/test
RUN rm -fr /usr/local/go/api
RUN rm -fr /usr/local/go/bin/godoc
RUN rm -fr /usr/local/go/bin/gofmt

# Build a hugo context
FROM ubuntu:xenial as build_hugo_context

ENV HUGO_VERSION=0.55.5
RUN mkdir -p /out
ADD https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_Linux-64bit.tar.gz /tmp
RUN tar -xzvf /tmp/hugo_${HUGO_VERSION}_Linux-64bit.tar.gz -C /out

# Build a nodejs tools context
FROM ubuntu:xenial as nodejs_tools_context
ENV NODEJS_VERSION=12.8.0

ADD https://nodejs.org/download/release/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-linux-x64.tar.gz /tmp
RUN tar -xzf /tmp/node-v${NODEJS_VERSION}-linux-x64.tar.gz --strip-components=1 -C /usr/local

ADD https://nodejs.org/download/release/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-headers.tar.gz /tmp
RUN tar -xzf /tmp/node-v${NODEJS_VERSION}-headers.tar.gz --strip-components=1 -C /usr/local

RUN npm init -y
RUN npm install -g \
    sass \
    sass-lint \
    typescript \
    tslint \
    markdown-spellcheck \
    svgstore-cli \
    svgo
RUN npm install --save-dev @babel/core @babel/cli @babel/preset-env
RUN npm install --save @babel/polyfill

# Build a ruby tools context
FROM ubuntu:xenial as ruby_tools_context
ENV RUBY_VERSION=2.6.3
RUN apt-get -y update
RUN apt-get -y install build-essential zlib1g-dev libssl-dev

ADD https://cache.ruby-lang.org/pub/ruby/2.6/ruby-${RUBY_VERSION}.tar.gz /tmp
WORKDIR /tmp
RUN tar -xzf ruby-${RUBY_VERSION}.tar.gz
WORKDIR /tmp/ruby-${RUBY_VERSION}
RUN ./configure
RUN make
RUN make install

# Install istio.io verification tools
RUN gem install --no-wrappers mdl
RUN gem install --no-wrappers html-proofer -v 3.9.2
RUN gem install --no-wrappers awesome_bot

# Build a base OS context
FROM ubuntu:xenial as base_os_context

# required for golang: ca-certificates, gcc, libc-dev, git
# required for general build: make, wget, curl
# required for ruby: libcurl4-openssl-dev
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    gcc \
    git \
    libc-dev \
    make \
    pkg-config \
    libcurl4-openssl-dev \
    wget

# Clean up stuff we don't need in the final image
RUN rm -rf /var/lib/apt/lists/*
RUN rm -fr /usr/share

# Prepare final output image
FROM scratch

# Setup to run the Go compiler with host-side build cache, test cache, and module cache
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org
ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV PATH=/go/bin:$PATH
ENV HOME=/root

# Uhh ya. Ruby.
ENV RUBYOPT="-KU -E utf-8:utf-8"

COPY --from=base_os_context / /
COPY --from=golang_context /usr/local/go/ /go/
COPY --from=build_hugo_context /out/ /usr/local/bin
COPY --from=build_tools_context /out/ /
COPY --from=ruby_tools_context /usr/local/bin /usr/local/bin
COPY --from=ruby_tools_context /usr/local/lib /usr/local/lib
COPY --from=nodejs_tools_context /usr/local /usr/local
COPY --from=nodejs_tools_context /package.json /

RUN mkdir -p /gocache && \
    mkdir -p /go/pkg/mod

# TODO must sort out how to use uid mapping in docker so these don't need to be 777
# They are created as root 755.  As a result they are not writeable, which fails in
# the developer environment as a volume or bind mount inherits the permissions of
# the directory mounted rather then overridding with the permission of the volume file.
RUN chmod 777 /gocache && \
    chmod 777 /go/pkg/mod && \
    chmod 777 /root && \
    chmod 777 /go/src

WORKDIR /

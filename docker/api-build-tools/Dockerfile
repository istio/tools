FROM golang:1.12 as golang_build_env

# Setup environment
ENV OUTDIR=/out
ENV PROTOBUF_VERSION=3.9.0
ENV GOGO_PROTOBUF_VERSION=1.2.1
ENV GOLANG_PROTOBUF_VERSION=1.3.1
ENV PROTOLOCK_VERSION=0.14.0
ENV PROTOTOOL_VERSION=1.8.0

RUN mkdir -p ${OUTDIR}/usr/bin
RUN mkdir -p ${OUTDIR}/usr/include

# Update distro and install dependencies
RUN apt-get update
RUN apt-get install -y build-essential unzip autoconf libtool zlibc zlib1g-dev libc-ares-dev libssl-dev upx

# Install protoc
WORKDIR /tmp
ADD https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOBUF_VERSION}/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip /tmp/
RUN unzip /tmp/protoc-${PROTOBUF_VERSION}-linux-x86_64.zip

# Instead of cloning the entire https://github.com/protobuf repo,
# copy only those protobufs needed for Istio's API builder
RUN mkdir -p /tmp/include/protobuf/google/protobuf
RUN mkdir -p /tmp/include/protobuf/google/rpc
RUN mkdir -p /tmp/include/protobuf/gogoproto
RUN for f in any duration descriptor empty struct timestamp wrappers; do \
            curl -L -o /tmp/include/protobuf/google/protobuf/${f}.proto https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/${f}.proto; \
        done
RUN for f in code error_details status http; do \
            curl -L -o /tmp/include/protobuf/google/rpc/${f}.proto https://raw.githubusercontent.com/istio/gogo-genproto/master/googleapis/google/rpc/${f}.proto; \
        done
RUN curl -L -o /tmp/include/protobuf/gogoproto/gogo.proto https://raw.githubusercontent.com/gogo/protobuf/master/gogoproto/gogo.proto

# Install protoc plugins
RUN GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gofast@v${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/gogo/protobuf/protoc-gen-gogoslick@v${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/uber/prototool@v${PROTOTOOL_VERSION}
RUN go get github.com/nilslice/protolock/...
# TODO(sdake) Not sure why a versioned version of protolock can't be retrieved.
#@v${PROTOLOCK_VERSION}

# Install custom istio.io tools
RUN go get istio.io/tools/protoc-gen-docs
RUN go get istio.io/tools/cmd/annotations_prep

# Install go static binaries
RUN install -c ${GOPATH}/bin/* ${OUTDIR}/usr/bin
RUN cp -aR /tmp/bin/protoc ${OUTDIR}/usr/bin
RUN cp -aR /tmp/include/* ${OUTDIR}/usr/include

# Optimization step
RUN upx --lzma /${OUTDIR}/usr/bin/*

# Finally, Render a distroless image with the binaries and protobuf includes
FROM gcr.io/distroless/base:latest

COPY --from=golang_build_env /out/ /

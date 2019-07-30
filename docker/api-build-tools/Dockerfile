FROM golang:1.12 as golang_build_env

# Setup environment
ENV OUTDIR=/out
ENV PROTOBUF_VERSION=3.6.1
ENV GOGO_PROTOBUF_VERSION=v1.2.2
ENV GOLANG_PROTOBUF_VERSION=v1.3.1
ENV PROTOLOCK_VERSION=v0.14.0
ENV PROTOTOOL_VERSION=v1.8.0
ENV COUNTERFEITER_VERSION=v6.2.2

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

# Install external dependencies
RUN GO111MODULE=on go get github.com/golang/protobuf/protoc-gen-go@${GOLANG_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/istio/gogo-protobuf/protoc-gen-gogofast@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/istio/gogo-protobuf/protoc-gen-gogofaster@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/istio/gogo-protobuf/protoc-gen-gogoslick@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get github.com/uber/prototool/cmd/prototool@${PROTOTOOL_VERSION}
RUN GO111MODULE=on go get github.com/nilslice/protolock/cmd/protolock@${PROTOLOCK_VERSION}


RUN GO111MODULE=on go get github.com/maxbrunsfeld/counterfeiter/v6@${COUNTERFEITER_VERSION}
RUN go get github.com/jteeuwen/go-bindata/...

# Install custom Istio tools
RUN go get istio.io/tools/protoc-gen-docs
RUN go get istio.io/tools/cmd/annotations_prep

# Put the stuff we need in its final output location
RUN install -c ${GOPATH}/bin/* ${OUTDIR}/usr/bin
RUN cp -aR /tmp/bin/protoc ${OUTDIR}/usr/bin
RUN cp -aR /tmp/include/* ${OUTDIR}/usr/include

# Optimization step
RUN upx --lzma /${OUTDIR}/usr/bin/*

# Finally, render the image with the binaries and protobuf includes
FROM ubuntu:xenial
COPY --from=golang_build_env /out/ /

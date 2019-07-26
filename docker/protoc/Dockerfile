FROM golang:1.12 as golang_build_env

# Setup environment
ENV OUTDIR=/out

# Update distro and install dependencies
RUN apt-get update
RUN apt-get install -y build-essential unzip autoconf libtool zlibc zlib1g-dev libc-ares-dev libssl-dev upx

# Install protoc
RUN curl -LO https://github.com/protocolbuffers/protobuf/releases/download/v3.9.0/protobuf-all-3.9.0.tar.gz
run tar -xzvf protobuf-all-3.9.0.tar.gz
WORKDIR protobuf-3.9.0
RUN aclocal
RUN automake
RUN ./autogen.sh
RUN ./configure --prefix=/usr
RUN make
RUN make install DESTDIR=${OUTDIR}
RUN make install

# Install gRPC
RUN curl -LO https://github.com/grpc/grpc/archive/v1.22.0.tar.gz
RUN tar -xzvf v1.22.0.tar.gz
WORKDIR grpc-1.22.0
RUN make install DESTDIR=${OUTDIR}

# Curl mysterious protobuf files
RUN mkdir -p /protobuf/google/protobuf
RUN mkdir -p /protobuf/google/rpc
RUN mkdir -p /protobuf/gogoproto
RUN for f in any duration descriptor empty struct timestamp wrappers; do \
            curl -L -o /protobuf/google/protobuf/${f}.proto https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/${f}.proto; \
        done
RUN for f in code error_details status http; do \
            curl -L -o /protobuf/google/rpc/${f}.proto https://raw.githubusercontent.com/istio/gogo-genproto/master/googleapis/google/rpc/${f}.proto; \
        done
RUN curl -L -o /protobuf/gogoproto/gogo.proto https://raw.githubusercontent.com/gogo/protobuf/master/gogoproto/gogo.proto

# Install protoc plugins
RUN go get -u -v -ldflags '-w -s' github.com/gogo/protobuf/protoc-gen-gogofast
RUN go get -u -v -ldflags '-w -s' github.com/gogo/protobuf/protoc-gen-gogofaster
RUN go get -u -v -ldflags '-w -s' github.com/gogo/protobuf/protoc-gen-gogoslick
RUN go get -u -v -ldflags '-w -s' istio.io/tools/protoc-gen-docs

# Install other tools needed by the API repo
RUN go get -u -v -ldflags '-w -s' istio.io/tools/cmd/annotations_prep

# Install go static binaries
RUN install -c ${GOPATH}/bin/* ${OUTDIR}/usr/bin

RUN upx -9 /out/usr/bin/*

# Finally, Render final Xenial distro based image
FROM ubuntu:xenial
RUN apt-get update
RUN apt-get install -y zlib1g zlibc libc-ares2

COPY --from=golang_build_env /out/ /
COPY --from=golang_build_env /protobuf/ /protobuf

ENTRYPOINT ["/usr/bin/protoc", "-I/protobuf"]

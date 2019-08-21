# Copyright 2019 Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This is used to build Istio's primary build container, which contains all the tools
# necessary to perform and all build activities in all Istio repos.
#
# The container is built using different contexts, one per execution environment (plain binaries, Ruby, nodejs), and
# then combined into the final image.
#
# We pin versions of stuff we install. Modify the various XXX_VERSION variables within the indivudual build contexts
# in order to control these versions.
#
# NOTE: This is still a work-in-progress and only some Istio repos use this. More tools will be
# added here as we onboard more repos.

#############
# Binary tools
#############

FROM golang:1.12.9 as binary_tools_context

# Pinned versions of stuff we pull in
ENV PROTOC_VERSION=3.6.1
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
ENV SHELLCHECK_VERSION=v0.7.0
ENV HUGO_VERSION=0.55.5
ENV HADOLINT_VERSION=v1.17.1
ENV UPX_VERSION=3.95
ENV HELM_VERSION=v2.10.0

ENV OUTDIR=/out
WORKDIR /tmp

# Update distro and install dependencies
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    unzip \
    xz-utils

# Install protoc
ADD https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip /tmp/
RUN unzip /tmp/protoc-${PROTOC_VERSION}-linux-x86_64.zip
RUN mkdir -p ${OUTDIR}/usr/bin
RUN mv /tmp/bin/protoc ${OUTDIR}/usr/bin

# Install necessary protoc includes
# TODO: Need to pin include file versions

RUN mkdir -p ${OUTDIR}/usr/include/protobuf/google/protobuf
RUN for f in any duration descriptor empty struct timestamp wrappers; do \
            curl -L -o ${OUTDIR}/usr/include/protobuf/google/protobuf/${f}.proto https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/${f}.proto; \
        done

RUN mkdir -p ${OUTDIR}/usr/include/protobuf/google/rpc
RUN for f in code error_details status; do \
            curl -L -o ${OUTDIR}/usr/include/protobuf/google/rpc/${f}.proto https://raw.githubusercontent.com/istio/gogo-genproto/master/googleapis/google/rpc/${f}.proto; \
        done

RUN mkdir -p ${OUTDIR}/usr/include/protobuf/google/api
RUN for f in http annotations; do \
            curl -L -o ${OUTDIR}/usr/include/protobuf/google/api/${f}.proto https://raw.githubusercontent.com/istio/gogo-genproto/master/googleapis/google/api/${f}.proto; \
        done

RUN mkdir -p ${OUTDIR}/usr/include/protobuf/validate
RUN curl -L -o ${OUTDIR}/usr/include/protobuf/validate/validate.proto https://raw.githubusercontent.com/envoyproxy/protoc-gen-validate/${PROTOC_GEN_VALIDATE_VERSION}/validate/validate.proto

RUN mkdir -p ${OUTDIR}/usr/include/protobuf/gogoproto
RUN curl -L -o ${OUTDIR}/usr/include/protobuf/gogoproto/gogo.proto https://raw.githubusercontent.com/gogo/protobuf/master/gogoproto/gogo.proto

RUN mkdir -p ${OUTDIR}/usr/include/protobuf/k8s.io/api
RUN git clone --depth 1 --single-branch --branch master https://github.com/kubernetes/api.git
WORKDIR /tmp/api
RUN find . -name \*proto -exec cp -a --parents {} ${OUTDIR}/usr/include/protobuf/k8s.io/api \;
WORKDIR /tmp

RUN mkdir -p ${OUTDIR}/usr/include/protobuf/k8s.io/apimachinery
RUN git clone --depth 1 --single-branch --branch master https://github.com/kubernetes/apimachinery.git
WORKDIR /tmp/apimachinery
RUN find . -name \*proto -exec cp -a --parents {} ${OUTDIR}/usr/include/protobuf/k8s.io/apimachinery \;
WORKDIR /tmp

RUN mkdir -p ${OUTDIR}/usr/include/protobuf/github.com/gogo/protobuf
RUN git clone --depth 1 --single-branch --branch master https://github.com/gogo/protobuf.git
WORKDIR /tmp/protobuf
RUN find . -name \*proto -exec cp -a --parents {} ${OUTDIR}/usr/include/protobuf/github.com/gogo/protobuf \;
RUN rm -fr ${OUTDIR}/usr/include/protobuf/github.com/gogo/protobuf/test
RUN rm -fr ${OUTDIR}/usr/include/protobuf/github.com/gogo/protobuf/protoc-gen-gogo/testdata
RUN rm -fr ${OUTDIR}/usr/include/protobuf/github.com/gogo/protobuf/jsonpb/jsonpb_test_proto
RUN rm -fr ${OUTDIR}/usr/include/protobuf/github.com/gogo/protobuf/conformance
RUN rm -fr ${OUTDIR}/usr/include/protobuf/github.com/gogo/protobuf/vanity/test
RUN rm -fr ${OUTDIR}/usr/include/protobuf/github.com/gogo/protobuf/proto/test_proto
WORKDIR /tmp

# Build and install a bunch of Go tools
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

# Compress the Go tools and put them in their final location
ADD https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-amd64_linux.tar.xz /tmp
RUN tar -xJf upx-${UPX_VERSION}-amd64_linux.tar.xz -C /tmp
RUN mv /tmp/upx-${UPX_VERSION}-amd64_linux/upx /usr/bin
RUN upx --lzma /go/bin/*
RUN mkdir -p ${OUTDIR}/go/bin
RUN mv /go/bin/* ${OUTDIR}/go/bin

# ShellCheck linter
ADD https://storage.googleapis.com/shellcheck/shellcheck-${SHELLCHECK_VERSION}.linux.x86_64.tar.xz /tmp
RUN tar -xJf /tmp/shellcheck-${SHELLCHECK_VERSION}.linux.x86_64.tar.xz -C /tmp
RUN mv /tmp/shellcheck-${SHELLCHECK_VERSION}/shellcheck ${OUTDIR}/usr/bin

# Hadolint linter
ADD https://github.com/hadolint/hadolint/releases/download/${HADOLINT_VERSION}/hadolint-Linux-x86_64 ${OUTDIR}/usr/bin/hadolint
RUN chmod 555 ${OUTDIR}/usr/bin/hadolint

# Hugo static site generator
ADD https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_Linux-64bit.tar.gz /tmp
RUN tar -xzvf /tmp/hugo_${HUGO_VERSION}_Linux-64bit.tar.gz -C /tmp
RUN mv /tmp/hugo ${OUTDIR}/usr/bin

# Helm
ADD https://storage.googleapis.com/kubernetes-helm/helm-${HELM_VERSION}-linux-amd64.tar.gz /tmp
RUN tar -xf /tmp/helm-${HELM_VERSION}-linux-amd64.tar.gz -C /tmp
RUN mv /tmp/linux-amd64/helm ${OUTDIR}/usr/bin

# Cleanup stuff we don't need in the final image
RUN rm -fr /usr/local/go/doc
RUN rm -fr /usr/local/go/test
RUN rm -fr /usr/local/go/api
RUN rm -fr /usr/local/go/bin/godoc
RUN rm -fr /usr/local/go/bin/gofmt

#############
# Nodejs
#############

FROM ubuntu:xenial as nodejs_tools_context

ENV NODEJS_VERSION=12.8.0
ENV MARKDOWN_SPELLCHECK_VERSION=v1.3.1
ENV SASS_LINT_VERSION=v1.13.1
ENV TYPESCRIPT_VERSION=v3.5.3
ENV TSLINT_VERSION=v5.18.0
ENV SASS_VERSION=v1.22.10
ENV SVGO_VERSION=v1.3.0
ENV SVGSTORE_CLI_VERSION=v1.3.1
ENV BABEL_CLI_VERSION=v7.5.5
ENV BABEL_CORE_VERSION=v7.5.5
ENV BABEL_PRESET_ENV=v7.5.5
ENV BABEL_PRESET_MINIFY_VERSION=v0.5.1
ENV BABEL_POLYFILL_VERSION=v7.4.4

ADD https://nodejs.org/download/release/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-linux-x64.tar.gz /tmp
RUN tar -xzf /tmp/node-v${NODEJS_VERSION}-linux-x64.tar.gz --strip-components=1 -C /usr/local

ADD https://nodejs.org/download/release/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-headers.tar.gz /tmp
RUN tar -xzf /tmp/node-v${NODEJS_VERSION}-headers.tar.gz --strip-components=1 -C /usr/local

RUN npm init -y
RUN npm install --production --global \
    sass@${SASS_VERSION} \
    sass-lint@${SASS_LINT_VERSION} \
    typescript@${TYPESCRIPT_VERSION} \
    tslint@${TSLINT_VERSION} \
    markdown-spellcheck@${MARKDONN_SPELLCHECK_VERSION} \
    svgstore-cli@${SVGSTORE_CLI_VERSION} \
    svgo@${SVGO_VERSION} \
    @babel/core@${BABEL_CORE_VERSION} \
    @babel/cli@${BABEL_CLI_VERSION} \
    @babel/preset-env@${BABEL_PRESET_ENV_VERSION}

RUN npm install --production --save-dev \
    babel-preset-minify@${BABEL_PRESET_MINIFY_VERSION}

RUN npm install --save-dev \
    @babel/polyfill@${BABEL_POLYFILL_VERSION}

# Clean up stuff we don't need in the final image
RUN rm -rf /usr/local/sbin
RUN rm -rf /usr/local/share

#############
# Ruby
#############

FROM ubuntu:xenial as ruby_tools_context

ENV RUBY_VERSION=2.6.3
ENV HTMLPROOFER_VERSION=3.9.2
ENV MDL_VERSION=0.5.0
ENV AWESOMEBOT_VERSION=1.19.1

RUN apt-get -y update && apt-get -y install --no-install-recommends \
    build-essential \
    libssl-dev \
    zlib1g-dev

ADD https://cache.ruby-lang.org/pub/ruby/2.6/ruby-${RUBY_VERSION}.tar.gz /tmp
WORKDIR /tmp
RUN tar -xzf ruby-${RUBY_VERSION}.tar.gz
WORKDIR /tmp/ruby-${RUBY_VERSION}
RUN ./configure
RUN make
RUN make install

# Install istio.io verification tools
RUN gem install --no-wrappers --no-document mdl -v ${MDL_VERSION}
RUN gem install --no-wrappers --no-document html-proofer -v ${HTMLPROOFER_VERSION}
RUN gem install --no-wrappers --no-document awesome_bot -v ${AWESOMEBOT_VERSION}

#############
# Python
#############

FROM ubuntu:xenial as python_context

ENV YAMLLINT_VERSION=1.17.0

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    libc-dev \
    pkg-config \
    python3

# Install pip
RUN curl "https://bootstrap.pypa.io/get-pip.py" -o "get-pip.py"
RUN python3 get-pip.py --no-wheel

# Install yamllint
RUN pip install yamllint==${YAMLLINT_VERSION}

# Clean up stuff we don't need in the final image
RUN rm -rf /usr/local/bin/pip*
RUN rm -fr /usr/local/lib/python3.5/dist-packages/pip

#############
# Base OS
#############

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
    libcurl4-openssl-dev \
    make \
    pkg-config \
    python3 \
    wget

# Clean up stuff we don't need in the final image
RUN rm -rf /var/lib/apt/lists/*
RUN rm -fr /usr/share

#############
# Final image
#############

# Prepare final output image
FROM scratch

# General
ENV HOME=/root

# Go support
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org
ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV PATH=/go/bin:$PATH

# Ruby suppport
ENV RUBYOPT="-KU -E utf-8:utf-8"

COPY --from=base_os_context / /
COPY --from=binary_tools_context /out/ /
COPY --from=binary_tools_context /usr/local/go/ /go/
COPY --from=ruby_tools_context /usr/local/bin /usr/local/bin
COPY --from=ruby_tools_context /usr/local/lib /usr/local/lib
COPY --from=nodejs_tools_context /usr/local /usr/local
COPY --from=nodejs_tools_context /node_modules/ /node_modules/
COPY --from=python_context /usr/local/bin /usr/local/bin
COPY --from=python_context /usr/local/lib /usr/local/lib

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

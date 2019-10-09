# Copyright Istio Authors
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

###############
# Binary tools
###############

FROM golang:1.13.1 as binary_tools_context

# Pinned versions of stuff we pull in
ENV COUNTERFEITER_VERSION=v6.2.2
ENV FORTIO_VERSION=v1.3.1
ENV GCR_AUTH_VERSION=1.5.0
ENV GO_BINDATA_VERSION=6025e8de665b31fa74ab1a66f2cddd8c0abf887e
ENV GO_JUNIT_REPORT_VERSION=af01ea7f8024089b458d804d5cdf190f962a9a0c
ENV GOCOVMERGE_VERSION=b5bfa59ec0adc420475f97f89b58045c721d761c
ENV GOGO_PROTOBUF_VERSION=v1.3.0
ENV GOIMPORTS_VERSION=379209517ffe
ENV GOLANG_PROTOBUF_VERSION=v1.3.1
ENV GOLANGCI_LINT_VERSION=v1.18.0
ENV HADOLINT_VERSION=v1.17.2
ENV HELM_VERSION=v2.10.0
ENV HUGO_VERSION=0.58.2
ENV JUNIT_MERGER_VERSION=adf1545b49509db1f83c49d1de90bbcb235642a8
ENV K8S_CODE_GENERATOR_VERSION=1.13.10
ENV KIND_VERSION=v0.5.1
ENV KUBECTL_VERSION=1.15.0
ENV PROTOC_GEN_GRPC_GATEWAY_VERSION=v1.8.1
ENV PROTOC_GEN_SWAGGER_VERSION=v1.8.1
ENV PROTOC_GEN_VALIDATE_VERSION=d6164de4910977d3c3c8dbd9299b5064ea9e7c2b
ENV PROTOC_VERSION=3.9.2
ENV PROTOLOCK_VERSION=v0.14.0
ENV PROTOTOOL_VERSION=7df3b957ffe3d09dc57fe4e1eb96694614db8c7a
ENV SHELLCHECK_VERSION=v0.7.0
ENV SU_EXEC_VERSION=0.2
ENV UPX_VERSION=3.95

WORKDIR /tmp
ENV GOPATH=/tmp/go

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/bin

# Update distro and install dependencies
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg2 \
    software-properties-common \
    unzip \
    xz-utils

# Install protoc
ADD https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip /tmp/
RUN unzip /tmp/protoc-${PROTOC_VERSION}-linux-x86_64.zip
RUN mv /tmp/bin/protoc ${OUTDIR}/usr/bin

# Build and install a bunch of Go tools
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/golang/protobuf/protoc-gen-go@${GOLANG_PROTOBUF_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/gogo/protobuf/protoc-gen-gofast@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/gogo/protobuf/protoc-gen-gogofast@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/gogo/protobuf/protoc-gen-gogofaster@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/gogo/protobuf/protoc-gen-gogoslick@${GOGO_PROTOBUF_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/uber/prototool/cmd/prototool@${PROTOTOOL_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/nilslice/protolock/cmd/protolock@${PROTOLOCK_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/maxbrunsfeld/counterfeiter/v6@${COUNTERFEITER_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/jteeuwen/go-bindata/go-bindata@${GO_BINDATA_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/envoyproxy/protoc-gen-validate@${PROTOC_GEN_VALIDATE_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@${PROTOC_GEN_GRPC_GATEWAY_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@${PROTOC_GEN_SWAGGER_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/jstemmer/go-junit-report@${GO_JUNIT_REPORT_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" sigs.k8s.io/kind@${KIND_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/wadey/gocovmerge@${GOCOVMERGE_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/fortio/fortio@${FORTIO_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" github.com/imsky/junit-merger/src/junit-merger@${JUNIT_MERGER_VERSION}

# Install latest version of Istio-owned tools
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/protoc-gen-docs
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/annotations_prep
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/cue-gen
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/envvarlinter
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/testlinter
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/protoc-gen-deepcopy
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/protoc-gen-jsonshim
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/kubetype-gen
RUN GO111MODULE=on go get -ldflags="-s -w" istio.io/tools/cmd/license-lint
RUN go get -ldflags="-s -w" -u istio.io/test-infra/toolbox/githubctl
RUN go get -ldflags="-s -w" -u istio.io/test-infra/boskos/cmd/mason_client

# Install k8s code generation tools
RUN GO111MODULE=on go get -ldflags="-s -w" k8s.io/code-generator/cmd/defaulter-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" k8s.io/code-generator/cmd/client-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" k8s.io/code-generator/cmd/lister-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" k8s.io/code-generator/cmd/informer-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" k8s.io/code-generator/cmd/deepcopy-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN GO111MODULE=on go get -ldflags="-s -w" k8s.io/code-generator/cmd/go-to-protobuf@kubernetes-${K8S_CODE_GENERATOR_VERSION}

# Compress the Go tools and put them in their final location
ADD https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-amd64_linux.tar.xz /tmp
RUN tar -xJf upx-${UPX_VERSION}-amd64_linux.tar.xz -C /tmp
RUN mv /tmp/upx-${UPX_VERSION}-amd64_linux/upx /usr/bin
RUN upx --lzma /tmp/go/bin/*
RUN mv /tmp/go/bin/* ${OUTDIR}/usr/bin

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

# Kubectl
ADD https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl ${OUTDIR}/usr/bin/kubectl
RUN chmod 555 ${OUTDIR}/usr/bin/kubectl

# GCR docker credential helper
ADD https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v${GCR_AUTH_VERSION}/docker-credential-gcr_linux_amd64-${GCR_AUTH_VERSION}.tar.gz /tmp
RUN tar -xzf /tmp/docker-credential-gcr_linux_amd64-${GCR_AUTH_VERSION}.tar.gz -C /tmp
RUN mv /tmp/docker-credential-gcr ${OUTDIR}/usr/bin

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

# Pinned versions of stuff we pull in
ENV BABEL_CLI_VERSION=v7.6.2
ENV BABEL_CORE_VERSION=v7.6.2
ENV BABEL_POLYFILL_VERSION=v7.4.4
ENV BABEL_PRESET_ENV=v7.6.2
ENV BABEL_PRESET_MINIFY_VERSION=v0.5.1
ENV MARKDOWN_SPELLCHECK_VERSION=v1.3.1
ENV NODEJS_VERSION=12.8.0
ENV SASS_LINT_VERSION=v1.13.1
ENV SASS_VERSION=v1.22.10
ENV SVGO_VERSION=v1.3.0
ENV SVGSTORE_CLI_VERSION=v1.3.1
ENV TSLINT_VERSION=v5.20.0
ENV TYPESCRIPT_VERSION=v3.6.3

ADD https://nodejs.org/download/release/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-linux-x64.tar.gz /tmp
RUN tar -xzf /tmp/node-v${NODEJS_VERSION}-linux-x64.tar.gz --strip-components=1 -C /usr/local

ADD https://nodejs.org/download/release/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-headers.tar.gz /tmp
RUN tar -xzf /tmp/node-v${NODEJS_VERSION}-headers.tar.gz --strip-components=1 -C /usr/local

RUN npm init -y
RUN npm install --production --global \
    sass@"${SASS_VERSION}" \
    sass-lint@"${SASS_LINT_VERSION}" \
    typescript@"${TYPESCRIPT_VERSION}" \
    tslint@"${TSLINT_VERSION}" \
    markdown-spellcheck@"${MARKDOWN_SPELLCHECK_VERSION}" \
    svgstore-cli@"${SVGSTORE_CLI_VERSION}" \
    svgo@"${SVGO_VERSION}" \
    @babel/core@"${BABEL_CORE_VERSION}" \
    @babel/cli@"${BABEL_CLI_VERSION}" \
    @babel/preset-env@"${BABEL_PRESET_ENV_VERSION}"

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

# Pinned versions of stuff we pull in
ENV AWESOMEBOT_VERSION=1.19.1
ENV FPM_VERSION=1.11.0
ENV HTMLPROOFER_VERSION=3.12.0
ENV LICENSEE_VERSION=9.11.0
ENV MDL_VERSION=0.5.0
ENV RUBY_VERSION=2.6.3

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    cmake \
    libssl-dev \
    pkg-config \
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
RUN gem install --no-wrappers --no-document licensee -v ${LICENSEE_VERSION}
RUN gem install --no-wrappers --no-document fpm -v ${FPM_VERSION}

#############
# Python
#############

FROM ubuntu:xenial as python_context

# Pinned versions of stuff we pull in
ENV AUTOPEP8_VERSION=1.4.4
ENV GCLOUD_VERSION=265.0.0
ENV YAMLLINT_VERSION=1.17.0

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    libc-dev \
    pkg-config \
    python3

# Install pip
RUN curl "https://bootstrap.pypa.io/get-pip.py" -o "get-pip.py"
RUN python3 get-pip.py --no-wheel

# Install Pyhton stuff
RUN pip install autopep8==${AUTOPEP8_VERSION}
RUN pip install yamllint==${YAMLLINT_VERSION}

# Clean up stuff we don't need in the final image
RUN rm -rf /usr/local/bin/pip*
RUN rm -fr /usr/local/lib/python3.5/dist-packages/pip

# TODO: the whole SDK is *huge*, but it's not clear what parts we need. We just want to be able to run auth-related gcloud
# commands, maybe we just need the gcloud binary out of all this?
ADD https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz /tmp
RUN tar -xzvf /tmp/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz -C /usr/local

#############
# Base OS
#############

FROM ubuntu:xenial as base_os_context

ENV DOCKER_VERSION=5:19.03.2~3-0~ubuntu-xenial
ENV CONTAINERD_VERSION=1.2.6-3

ENV OUTDIR=/out

# required for binary tools: ca-certificates, gcc, libc-dev, git, iptables, libltdl7
# required for general build: make, wget, curl
# required for ruby: libcurl4-openssl-dev
# required for python: python3, pkg-config
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common \
    ca-certificates \
    gcc \
    git \
    iptables \
    libltdl7 \
    libc-dev \
    libcurl4-openssl-dev \
    make \
    pkg-config \
    python3 \
    daemon \
    wget

# Docker including docker-ce, docker-ce-cli, and containerd.io
ADD https://download.docker.com/linux/ubuntu/gpg /tmp/docker-key
RUN apt-key add /tmp/docker-key
RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu xenial stable"
RUN apt-get update
RUN apt-get -y install --no-install-recommends docker-ce="${DOCKER_VERSION}" docker-ce-cli="${DOCKER_VERSION}" containerd.io="${CONTAINERD_VERSION}"

# Clean up stuff we don't need in the final image
RUN rm -rf /var/lib/apt/lists/*
RUN rm -fr /usr/share/perl
RUN rm -fr /usr/share/perl5
RUN rm -fr /usr/share/python
RUN rm -fr /usr/share/bash-completion
RUN rm -fr /usr/share/bug
RUN rm -fr /usr/share/doc
RUN rm -fr /usr/share/dh-python
RUN rm -fr /usr/share/locale
RUN rm -fr /usr/share/man
RUN rm -fr /tmp/*

# Run dockerd in CI
COPY prow-entrypoint.sh /usr/local/bin/entrypoint
RUN chmod +x /usr/local/bin/entrypoint

##############
# Final image
##############

# Prepare final output image
FROM scratch

# General
ENV HOME=/home
ENV LANG=C.UTF-8

# Go support
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org
ENV GOSUMDB=sum.golang.org
ENV GOROOT=/usr/local/go
ENV GOPATH=/go
ENV GOCACHE=/gocache
ENV GOBIN=/gobin
ENV PATH=/usr/local/go/bin:/gobin:/usr/local/google-cloud-sdk/bin:$PATH

# Ruby suppport
ENV RUBYOPT="-KU -E utf-8:utf-8"

# Create the file system
COPY --from=base_os_context / /
COPY --from=binary_tools_context /out/ /
COPY --from=binary_tools_context /usr/local/go /usr/local/go
COPY --from=ruby_tools_context /usr/local/bin /usr/local/bin
COPY --from=ruby_tools_context /usr/local/lib /usr/local/lib
COPY --from=nodejs_tools_context /usr/local /usr/local
COPY --from=nodejs_tools_context /node_modules /node_modules
COPY --from=python_context /usr/local/bin /usr/local/bin
COPY --from=python_context /usr/local/lib /usr/local/lib
COPY --from=python_context /usr/local/google-cloud-sdk /usr/local/google-cloud-sdk

COPY bashrc /home/.bashrc

RUN mkdir -p /go && \
    mkdir -p /gocache && \
    mkdir -p /gobin && \
    mkdir -p /home/.kube && \
    mkdir -p /home/.config

# TODO must sort out how to use uid mapping in docker so these don't need to be 777
# They are created as root 755.  As a result they are not writeable, which fails in
# the developer environment as a volume or bind mount inherits the permissions of
# the directory mounted rather then overridding with the permission of the volume file.
RUN chmod 777 /go && \
    chmod 777 /gocache && \
    chmod 777 /gobin && \
    chmod 777 /home/.kube && \
    chmod 777 /home/.config

WORKDIR /

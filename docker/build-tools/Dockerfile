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
# We pin versions of stuff we install. Modify the various XXX_VERSION variables within the individual build contexts
# in order to control these versions.

################
# Binary tools
################
ARG GOLANG_IMAGE=golang:1.18.4
# hadolint ignore=DL3006
FROM ${GOLANG_IMAGE} as binary_tools_context
# TARGETARCH is an automatic platform ARG enabled by Docker BuildKit.
ARG TARGETARCH

# Istio tools SHA that we use for this image
ARG ISTIO_TOOLS_SHA

# Pinned versions of stuff we pull in
ENV BOM_VERSION=v0.2.2
ENV GCR_AUTH_VERSION=2.0.4
ENV GO_BINDATA_VERSION=v3.1.2
ENV GO_JUNIT_REPORT_VERSION=df0ed838addb0fa189c4d76ad4657f6007a5811c
ENV GOCOVMERGE_VERSION=b5bfa59ec0adc420475f97f89b58045c721d761c
ENV GOIMPORTS_VERSION=v0.1.0
ENV BENCHSTAT_VERSION=9c9101da8316
ENV GH_VERSION=2.14.1
ENV GOLANG_PROTOBUF_VERSION=v1.28.0
ENV GOLANG_GRPC_PROTOBUF_VERSION=v1.2.0
# When updating the golangci version, you may want to update the common-files config/.golangci* files as well.
ENV GOLANGCI_LINT_VERSION=v1.46.1
ENV HADOLINT_VERSION=v2.10.0
ENV HELM3_VERSION=v3.9.0
ENV HUGO_VERSION=0.97.3
ENV JB_VERSION=v0.3.1
ENV JSONNET_VERSION=v0.15.0
ENV JUNIT_MERGER_VERSION=adf1545b49509db1f83c49d1de90bbcb235642a8
ENV K8S_CODE_GENERATOR_VERSION=1.24.1
ENV K8S_TEST_INFRA_VERSION=2acdc6800510dd422bfd2a5d8c02aedc19d15f8d
ENV KIND_VERSION=v0.14.0
ENV KUBECTL_VERSION=1.24.2
ENV PROTOC_GEN_GRPC_GATEWAY_VERSION=v1.8.6
ENV PROTOC_GEN_SWAGGER_VERSION=v1.8.6
ENV PROTOC_GEN_VALIDATE_VERSION=v0.6.7
ENV PROTOC_VERSION=21.2
ENV PROTOLOCK_VERSION=v0.14.0
ENV PROTOTOOL_VERSION=v1.10.0
ENV SHELLCHECK_VERSION=v0.8.0
ENV SU_EXEC_VERSION=0.2
ENV UPX_VERSION=3.96
ENV YQ_VERSION=4.25.3
ENV KPT_VERSION=v0.39.3
ENV BUF_VERSION=v1.5.0
ENV GCLOUD_VERSION=377.0.0
ENV KUBETEST2_VERSION=b019714a389563c9a788f119f801520d059b6533
ENV COSIGN_VERSION=v1.9.0
ENV CRANE_VERSION=v0.10.0

ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org

WORKDIR /tmp
ENV GOPATH=/tmp/go

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/bin
RUN mkdir -p ${OUTDIR}/usr/local

# Update distro and install dependencies
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    build-essential \
    ca-certificates \
    curl \
    gnupg2 \
    software-properties-common \
    unzip \
    xz-utils

# Install protoc
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) export PROTOC_ZIP=protoc-${PROTOC_VERSION}-linux-x86_64.zip;; \
        aarch64) export PROTOC_ZIP=protoc-${PROTOC_VERSION}-linux-aarch_64.zip;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    \
    wget -nv -O "/tmp/${PROTOC_ZIP}" "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP}"; \
    unzip "/tmp/${PROTOC_ZIP}"; \
    mv /tmp/bin/protoc ${OUTDIR}/usr/bin; \
    chmod +x ${OUTDIR}/usr/bin/protoc

# Install gh
ADD https://github.com/cli/cli/releases/download/v${GH_VERSION}/gh_${GH_VERSION}_linux_${TARGETARCH}.deb /tmp/
RUN dpkg -i /tmp/gh_${GH_VERSION}_linux_${TARGETARCH}.deb
RUN mv /usr/bin/gh ${OUTDIR}/usr/bin

# Build and install a bunch of Go tools
RUN go install -ldflags="-s -w" google.golang.org/protobuf/cmd/protoc-gen-go@${GOLANG_PROTOBUF_VERSION}
RUN go install -ldflags="-s -w" google.golang.org/grpc/cmd/protoc-gen-go-grpc@${GOLANG_GRPC_PROTOBUF_VERSION}
RUN go install -ldflags="-s -w" github.com/uber/prototool/cmd/prototool@${PROTOTOOL_VERSION}
RUN go install -ldflags="-s -w" github.com/nilslice/protolock/cmd/protolock@${PROTOLOCK_VERSION}
RUN go install -ldflags="-s -w" golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}
RUN go install -ldflags="-s -w" github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
RUN go install -ldflags="-s -w" github.com/go-bindata/go-bindata/go-bindata@${GO_BINDATA_VERSION}
RUN go install -ldflags="-s -w" github.com/envoyproxy/protoc-gen-validate@${PROTOC_GEN_VALIDATE_VERSION}
RUN go install -ldflags="-s -w" github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@${PROTOC_GEN_GRPC_GATEWAY_VERSION}
RUN go install -ldflags="-s -w" github.com/google/go-jsonnet/cmd/jsonnet@${JSONNET_VERSION}
RUN go install -ldflags="-s -w" github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@${JB_VERSION}
RUN go install -ldflags="-s -w" github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@${PROTOC_GEN_SWAGGER_VERSION}
RUN go install -ldflags="-s -w" github.com/istio/go-junit-report@${GO_JUNIT_REPORT_VERSION}
RUN go install -ldflags="-s -w" sigs.k8s.io/bom/cmd/bom@${BOM_VERSION}
RUN go install -ldflags="-s -w" sigs.k8s.io/kind@${KIND_VERSION}
RUN go install -ldflags="-s -w" github.com/wadey/gocovmerge@${GOCOVMERGE_VERSION}
RUN go install -ldflags="-s -w" github.com/imsky/junit-merger/src/junit-merger@${JUNIT_MERGER_VERSION}
RUN go install -ldflags="-s -w" golang.org/x/perf/cmd/benchstat@${BENCHSTAT_VERSION}
RUN go install -ldflags="-s -w" github.com/google/go-containerregistry/cmd/crane@${CRANE_VERSION}

# Install latest version of Istio-owned tools in this release
RUN go install -ldflags="-s -w" istio.io/tools/cmd/protoc-gen-docs@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/annotations_prep@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/cue-gen@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/envvarlinter@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/testlinter@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/protoc-gen-golang-deepcopy@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/protoc-gen-golang-jsonshim@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/kubetype-gen@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/license-lint@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" istio.io/tools/cmd/gen-release-notes@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" k8s.io/code-generator/cmd/applyconfiguration-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN go install -ldflags="-s -w" k8s.io/code-generator/cmd/defaulter-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN go install -ldflags="-s -w" k8s.io/code-generator/cmd/client-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN go install -ldflags="-s -w" k8s.io/code-generator/cmd/lister-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN go install -ldflags="-s -w" k8s.io/code-generator/cmd/informer-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN go install -ldflags="-s -w" k8s.io/code-generator/cmd/deepcopy-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN go install -ldflags="-s -w" k8s.io/code-generator/cmd/go-to-protobuf@kubernetes-${K8S_CODE_GENERATOR_VERSION}

# Install istio/test-infra tools
RUN go install sigs.k8s.io/kubetest2@${KUBETEST2_VERSION}
RUN go install sigs.k8s.io/kubetest2/kubetest2-gke@${KUBETEST2_VERSION}
RUN go install sigs.k8s.io/kubetest2/kubetest2-tester-exec@${KUBETEST2_VERSION}

# Go doesn't like the `replace` directives; need to do manual cloning now.
# Should be fixed by https://github.com/kubernetes/test-infra/issues/20421
# hadolint ignore=DL3003
RUN git clone https://github.com/kubernetes/test-infra --branch master --single-branch && \
  cd test-infra && \
  git checkout "${K8S_TEST_INFRA_VERSION}" && \
  go install ./robots/pr-creator && \
  go install ./prow/cmd/peribolos && \
  go install ./pkg/benchmarkjunit && \
  cd .. && rm -rf test-infra


# Compress the Go tools and put them in their final location
ADD https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-${TARGETARCH}_linux.tar.xz /tmp
RUN tar -xJf upx-${UPX_VERSION}-${TARGETARCH}_linux.tar.xz -C /tmp
RUN mv /tmp/upx-${UPX_VERSION}-${TARGETARCH}_linux/upx /usr/bin
RUN upx --lzma /tmp/go/bin/*
RUN mv /tmp/go/bin/* ${OUTDIR}/usr/bin

# Add gen-release-notes templates to filesystem
RUN mkdir -p ${OUTDIR}/usr/share/gen-release-notes
ADD https://raw.githubusercontent.com/istio/tools/master/cmd/gen-release-notes/templates/minorReleaseNotes.md ${OUTDIR}/usr/share/gen-release-notes
ADD https://raw.githubusercontent.com/istio/tools/master/cmd/gen-release-notes/templates/releaseNotes.md ${OUTDIR}/usr/share/gen-release-notes
ADD https://raw.githubusercontent.com/istio/tools/master/cmd/gen-release-notes/templates/upgradeNotes.md ${OUTDIR}/usr/share/gen-release-notes
RUN chmod -R 555 ${OUTDIR}/usr/share/gen-release-notes

# ShellCheck linter
RUN wget -nv -O "/tmp/shellcheck-${SHELLCHECK_VERSION}.linux.$(uname -m).tar.xz" "https://github.com/koalaman/shellcheck/releases/download/${SHELLCHECK_VERSION}/shellcheck-${SHELLCHECK_VERSION}.linux.$(uname -m).tar.xz"
RUN tar -xJf "/tmp/shellcheck-${SHELLCHECK_VERSION}.linux.$(uname -m).tar.xz" -C /tmp
RUN mv /tmp/shellcheck-${SHELLCHECK_VERSION}/shellcheck ${OUTDIR}/usr/bin

# Hadolint linter
ADD https://github.com/hadolint/hadolint/releases/download/${HADOLINT_VERSION}/hadolint-Linux-x86_64 ${OUTDIR}/usr/bin/hadolint
RUN chmod 555 ${OUTDIR}/usr/bin/hadolint

# Hugo static site generator
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) export HUGO_TAR=hugo_${HUGO_VERSION}_Linux-64bit.tar.gz;; \
        aarch64) export HUGO_TAR=hugo_${HUGO_VERSION}_Linux-ARM64.tar.gz;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    \
    wget -nv -O /tmp/${HUGO_TAR} https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${HUGO_TAR}; \
    tar -xzvf /tmp/${HUGO_TAR} -C /tmp; \
    mv /tmp/hugo ${OUTDIR}/usr/bin

# Helm version 3
ADD https://get.helm.sh/helm-${HELM3_VERSION}-linux-${TARGETARCH}.tar.gz /tmp
RUN mkdir /tmp/helm3
RUN tar -xf /tmp/helm-${HELM3_VERSION}-linux-${TARGETARCH}.tar.gz -C /tmp/helm3
RUN mv /tmp/helm3/linux-${TARGETARCH}/helm ${OUTDIR}/usr/bin/helm3
RUN ln ${OUTDIR}/usr/bin/helm3 ${OUTDIR}/usr/bin/helm

# yq doesn't support go modules, so install the binary instead
ADD https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_linux_${TARGETARCH} /tmp
RUN mv /tmp/yq_linux_${TARGETARCH} ${OUTDIR}/usr/bin/yq
RUN chmod 555 ${OUTDIR}/usr/bin/yq

# Kubectl
ADD https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl ${OUTDIR}/usr/bin/kubectl
RUN chmod 555 ${OUTDIR}/usr/bin/kubectl

# GCR docker credential helper
ADD https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v${GCR_AUTH_VERSION}/docker-credential-gcr_linux_${TARGETARCH}-${GCR_AUTH_VERSION}.tar.gz /tmp
RUN tar -xzf /tmp/docker-credential-gcr_linux_${TARGETARCH}-${GCR_AUTH_VERSION}.tar.gz -C /tmp
RUN mv /tmp/docker-credential-gcr ${OUTDIR}/usr/bin

RUN wget -nv -O "${OUTDIR}/usr/bin/buf" "https://github.com/bufbuild/buf/releases/download/${BUF_VERSION}/buf-Linux-$(uname -m)" && \
    chmod 555 "${OUTDIR}/usr/bin/buf"

# Install su-exec which is a tool that operates like sudo without the overhead
ADD https://github.com/ncopa/su-exec/archive/v${SU_EXEC_VERSION}.tar.gz /tmp
RUN tar -xzvf v${SU_EXEC_VERSION}.tar.gz
WORKDIR /tmp/su-exec-${SU_EXEC_VERSION}
RUN make
RUN cp -a su-exec ${OUTDIR}/usr/bin

ADD https://github.com/GoogleContainerTools/kpt/releases/download/${KPT_VERSION}/kpt_linux_${TARGETARCH} ${OUTDIR}/usr/bin/kpt
RUN chmod 555 ${OUTDIR}/usr/bin/kpt

# Install gcloud command line tool
# Install gcloud beta component
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64)  export GCLOUD_TAR_FILE="google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz" ;; \
        aarch64) export GCLOUD_TAR_FILE="google-cloud-sdk-${GCLOUD_VERSION}-linux-arm.tar.gz" ;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    \
    wget -nv "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/${GCLOUD_TAR_FILE}"; \
    tar -xzvf ."/${GCLOUD_TAR_FILE}" -C "${OUTDIR}/usr/local" && rm "${GCLOUD_TAR_FILE}"; \
    ${OUTDIR}/usr/local/google-cloud-sdk/bin/gcloud components install beta --quiet; \
    ${OUTDIR}/usr/local/google-cloud-sdk/bin/gcloud components install alpha --quiet; \
    rm -rf ${OUTDIR}/usr/local/google-cloud-sdk/.install/.backup \
    rm -rf ${OUTDIR}/usr/local/google-cloud-sdk/bin/anthoscli

# Install cosign (for signing build artifacts) and verify signature
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN set -eux; \
    ${OUTDIR}/usr/local/google-cloud-sdk/bin/gsutil -q cp gs://cosign-releases/${COSIGN_VERSION}/cosign-linux-${TARGETARCH} /tmp/cosign \
    && ${OUTDIR}/usr/local/google-cloud-sdk/bin/gsutil -q cat gs://cosign-releases/${COSIGN_VERSION}/cosign-linux-${TARGETARCH}.sig | base64 -d > /tmp/cosign.sig \
    && wget -nv -O /tmp/cosign-pubkey https://raw.githubusercontent.com/sigstore/cosign/main/release/release-cosign.pub \
    && openssl dgst -sha256 -verify /tmp/cosign-pubkey -signature /tmp/cosign.sig /tmp/cosign \
    && chmod +x /tmp/cosign \
    && mv /tmp/cosign ${OUTDIR}/usr/bin/ || exit 1

# Cleanup stuff we don't need in the final image
RUN rm -fr /usr/local/go/doc
RUN rm -fr /usr/local/go/test
RUN rm -fr /usr/local/go/api
RUN rm -fr /usr/local/go/bin/godoc
RUN rm -fr /usr/local/go/bin/gofmt

#############
# Node.js
#############
FROM ubuntu:jammy as nodejs_tools_context

WORKDIR /node

# Pinned versions of stuff we pull in
ENV BABEL_CLI_VERSION=v7.17.10
ENV BABEL_CORE_VERSION=v7.18.2
ENV BABEL_POLYFILL_VERSION=v7.12.1
ENV BABEL_PRESET_ENV=v7.18.2
ENV BABEL_PRESET_MINIFY_VERSION=v0.5.2
ENV LINKINATOR_VERSION=v2.0.5
ENV MARKDOWN_SPELLCHECK_VERSION=v1.3.1
ENV NODEJS_VERSION=18.2.0
ENV SASS_LINT_VERSION=v1.13.1
ENV SASS_VERSION=v1.52.1
ENV SVGO_VERSION=v1.3.2
ENV SVGSTORE_CLI_VERSION=v1.3.2
ENV TSLINT_VERSION=v6.1.3
ENV TYPESCRIPT_VERSION=v4.7.2

RUN apt-get update && apt-get install -y --no-install-recommends \
    wget ca-certificates

RUN set -eux; \
    case $(uname -m) in \
        x86_64) export NODEJS_TAR=node-v${NODEJS_VERSION}-linux-x64.tar.gz;; \
        aarch64) export NODEJS_TAR=node-v${NODEJS_VERSION}-linux-arm64.tar.gz;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    wget -nv -O /tmp/${NODEJS_TAR} https://nodejs.org/download/release/v${NODEJS_VERSION}/${NODEJS_TAR}; \
    tar -xzf /tmp/${NODEJS_TAR} --strip-components=1 -C /usr/local

ADD https://nodejs.org/download/release/v${NODEJS_VERSION}/node-v${NODEJS_VERSION}-headers.tar.gz /tmp
RUN tar -xzf /tmp/node-v${NODEJS_VERSION}-headers.tar.gz --strip-components=1 -C /usr/local

RUN npm init -y
RUN npm install --omit=dev --global \
    sass@"${SASS_VERSION}" \
    sass-lint@"${SASS_LINT_VERSION}" \
    typescript@"${TYPESCRIPT_VERSION}" \
    tslint@"${TSLINT_VERSION}" \
    markdown-spellcheck@"${MARKDOWN_SPELLCHECK_VERSION}" \
    svgstore-cli@"${SVGSTORE_CLI_VERSION}" \
    svgo@"${SVGO_VERSION}" \
    @babel/core@"${BABEL_CORE_VERSION}" \
    @babel/cli@"${BABEL_CLI_VERSION}" \
    @babel/preset-env@"${BABEL_PRESET_ENV_VERSION}" \
    linkinator@"${LINKINATOR_VERSION}"

RUN npm install --omit=dev --save-dev \
    babel-preset-minify@${BABEL_PRESET_MINIFY_VERSION}

RUN npm install --save-dev \
    @babel/polyfill@${BABEL_POLYFILL_VERSION}

# Clean up stuff we don't need in the final image
RUN rm -rf /usr/local/sbin
RUN rm -rf /usr/local/share

#############
# Ruby
#############

FROM ubuntu:jammy as ruby_tools_context

ENV DEBIAN_FRONTEND=noninteractive

# Pinned versions of stuff we pull in
ENV AWESOMEBOT_VERSION=1.20.0
# Sometime between 1.14.1 and 1.14.2 to pull in some fixes for deb
ENV FPM_VERSION=eb5370d16e361db3f1425f8c898bafe7f3c66869
ENV HTMLPROOFER_VERSION=3.19.0
ENV LICENSEE_VERSION=9.15.1
ENV MDL_VERSION=0.11.0

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    gnupg2 \
    software-properties-common \
    build-essential \
    zlib1g-dev \
    cmake \
    pkg-config \
    libssl-dev \
    git

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    ruby3.0 \
    ruby3.0-dev

# Install istio.io verification tools
RUN gem install --no-wrappers --no-document mdl -v ${MDL_VERSION}
RUN gem install --no-wrappers --no-document html-proofer -v ${HTMLPROOFER_VERSION}
RUN gem install --no-wrappers --no-document awesome_bot -v ${AWESOMEBOT_VERSION}
RUN gem install --no-wrappers --no-document licensee -v ${LICENSEE_VERSION}
# hadolint ignore=DL3003,DL3028
RUN git clone https://github.com/jordansissel/fpm && \
    cd fpm && \
    git reset --hard ${FPM_VERSION} && \
    make install

##############
# Python
##############

FROM ubuntu:jammy as python_context

ENV DEBIAN_FRONTEND=noninteractive

# Pinned versions of stuff we pull in
ENV AUTOPEP8_VERSION=1.4.4
ENV YAMLLINT_VERSION=1.24.2
ENV PIP_INSTALL_VERSION=21.0.1
ENV REQUESTS_VERSION=2.22.0
ENV PYTHON_PROTOBUF_VERSION=3.11.2
ENV PYYAML_VERSION=5.3.1
ENV JWCRYPTO_VERSION=0.7

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    libc-dev \
    pkg-config \
    python3 \
    python3-distutils \
    python3-pip \
    python3-setuptools

# Install Python stuff
RUN python3 -m pip install --no-cache-dir --upgrade pip==${PIP_INSTALL_VERSION}
RUN python3 -m pip install --no-cache-dir --no-binary :all: autopep8==${AUTOPEP8_VERSION}
RUN python3 -m pip install --no-cache-dir yamllint==${YAMLLINT_VERSION}
RUN python3 -m pip install --no-cache-dir requests==${REQUESTS_VERSION}
RUN python3 -m pip install --no-cache-dir protobuf==${PYTHON_PROTOBUF_VERSION}
RUN python3 -m pip install --no-cache-dir PyYAML==${PYYAML_VERSION}
RUN python3 -m pip install --no-cache-dir jwcrypto==${JWCRYPTO_VERSION}

#############
# Base OS
#############

FROM ubuntu:jammy as base_os_context

ENV DEBIAN_FRONTEND=noninteractive

ENV DOCKER_VERSION=5:20.10.14~3-0~ubuntu-jammy
ENV CONTAINERD_VERSION=1.5.11-1
ENV TRIVY_VERSION=0.28.1

ENV OUTDIR=/out

# required for binary tools: ca-certificates, gcc, libc-dev, git, iptables, libltdl7, less
# required for general build: make, wget, curl, ssh, rpm
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
    ssh \
    iptables \
    libltdl7 \
    libc-dev \
    libcurl4-openssl-dev \
    less \
    make \
    pkg-config \
    python3 \
    python3-setuptools \
    daemon \
    wget \
    rpm \
    jq \
    gettext-base \
    locales-all \
    file

# Fix Docker issue
RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

# Docker including docker-ce, docker-ce-cli, and containerd.io
ADD https://download.docker.com/linux/ubuntu/gpg /tmp/docker-key
RUN apt-key add /tmp/docker-key
ARG TARGETARCH
RUN add-apt-repository "deb [arch=${TARGETARCH}] https://download.docker.com/linux/ubuntu $(lsb_release -sc) stable"
RUN apt-get update
RUN apt-get -y install --no-install-recommends docker-ce="${DOCKER_VERSION}" docker-ce-cli="${DOCKER_VERSION}" containerd.io="${CONTAINERD_VERSION}"

# Trivy container scanner
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) \
            export TRVIY_DEB_NAME="trivy_${TRIVY_VERSION}_Linux-64bit.deb"; \
            ;; \
        aarch64) \
            export TRVIY_DEB_NAME="trivy_${TRIVY_VERSION}_Linux-ARM64.deb"; \
            ;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    wget -nv -O "/tmp/${TRVIY_DEB_NAME}" "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/${TRVIY_DEB_NAME}"; \
    apt-get -y install --no-install-recommends -f "/tmp/${TRVIY_DEB_NAME}"; \
    rm "/tmp/${TRVIY_DEB_NAME}";

# Clean up stuff we don't need in the final image
RUN rm -rf /var/lib/apt/lists/*
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

# Run config setup in local environments
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint

##############
# Final image
##############

# Prepare final output image
FROM scratch as build_tools

# Version from build arguments
ARG VERSION

# Labels used by Docker
LABEL "io.istio.repo"="https://github.com/istio/tools"
LABEL "io.istio.version"="${VERSION}"

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

# Ruby support
ENV RUBYOPT="-KU -E utf-8:utf-8"

# Create the file system
COPY --from=base_os_context / /
COPY --from=binary_tools_context /out/ /
COPY --from=binary_tools_context /usr/local/go /usr/local/go

COPY --from=nodejs_tools_context /usr/local/bin /usr/local/bin
COPY --from=nodejs_tools_context /usr/local/lib/node_modules /usr/local/lib/node_modules
COPY --from=nodejs_tools_context /node/node_modules /node_modules

COPY --from=ruby_tools_context /usr/bin /usr/bin
COPY --from=ruby_tools_context /usr/lib /usr/lib
COPY --from=ruby_tools_context /etc/alternatives /etc/alternatives
COPY --from=ruby_tools_context /var/lib/gems /var/lib/gems
COPY --from=ruby_tools_context /usr/local/bin /usr/local/bin

COPY --from=python_context /usr/local/bin /usr/local/bin
COPY --from=python_context /usr/local/lib /usr/local/lib

# su-exec is used in place of complex sudo setup operations
RUN chmod u+sx /usr/bin/su-exec

COPY bashrc /home/.bashrc

# mountpoints are mandatory for any host mounts.
# mountpoints in /config are special.
RUN mkdir -p /go && \
    mkdir -p /gocache && \
    mkdir -p /gobin && \
    mkdir -p /config/.docker && \
    mkdir -p /config/.config/gcloud && \
    mkdir -p /config/.kube && \
    mkdir -p /config-copy && \
    mkdir -p /home/.cache && \
    mkdir -p /home/.helm && \
    mkdir -p /home/.gsutil

# TODO must sort out how to use uid mapping in docker so these don't need to be 777
# They are created as root 755.  As a result they are not writeable, which fails in
# the developer environment as a volume or bind mount inherits the permissions of
# the directory mounted rather then overriding with the permission of the volume file.
RUN chmod 777 /go && \
    chmod 777 /gocache && \
    chmod 777 /gobin && \
    chmod 777 /config && \
    chmod 777 /config/.docker && \
    chmod 777 /config/.config/gcloud && \
    chmod 777 /config/.kube && \
    chmod 777 /home/.cache && \
    chmod 777 /home/.helm && \
    chmod 777 /home/.gsutil

WORKDIR /

ENTRYPOINT ["/usr/local/bin/docker-entrypoint"]


##############
# Clang+LLVM
##############

FROM ubuntu:xenial AS clang_context_amd64
FROM ubuntu:bionic AS clang_context_arm64
# hadolint ignore=DL3006
FROM clang_context_${TARGETARCH} AS clang_context

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    xz-utils \
    wget \
    ca-certificates

# 12.0.1 is the version support ubuntu:xenial & aarch64
ENV LLVM_VERSION=12.0.1
ENV LLVM_BASE_URL=https://github.com/llvm/llvm-project/releases/download/llvmorg-${LLVM_VERSION}
ENV LLVM_DIRECTORY=/usr/lib/llvm

RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) \
               export LLVM_ARCHIVE=clang+llvm-${LLVM_VERSION}-x86_64-linux-gnu-ubuntu- \
               export LLVM_ARTIFACT=clang+llvm-${LLVM_VERSION}-x86_64-linux-gnu-ubuntu-16.04;; \
        aarch64)  \
               export LLVM_ARCHIVE=clang+llvm-${LLVM_VERSION}-aarch64-linux-gnu \
               export LLVM_ARTIFACT=clang+llvm-${LLVM_VERSION}-aarch64-linux-gnu;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    \
    wget -nv ${LLVM_BASE_URL}/${LLVM_ARTIFACT}.tar.xz; \
    tar -xJf ${LLVM_ARTIFACT}.tar.xz -C /tmp; \
    mkdir -p ${LLVM_DIRECTORY}; \
    mv /tmp/${LLVM_ARCHIVE}/* ${LLVM_DIRECTORY}/

###########
# GN
###########

FROM debian:buster AS gn_context

RUN set -eux; \
    \
    apt-get update; \
    apt-get install -qqy --no-install-recommends \
        ca-certificates git \
        clang python ninja-build \
        libclang-dev libc++-dev \
        ;

WORKDIR /tmp
RUN git clone https://gn.googlesource.com/gn;

WORKDIR /tmp/gn

RUN set -eux; \
    \
    git checkout 501b49a3; \
    python build/gen.py; \
    ninja -v -C out; \
    out/gn_unittests; \
    mkdir -p /gn; \
    cp /tmp/gn/out/gn /gn/gn; \
    /gn/gn --version;

###########
# Bazel
###########

FROM ubuntu:xenial AS bazel_context_amd64
FROM ubuntu:bionic AS bazel_context_arm64
# hadolint ignore=DL3006
FROM bazel_context_${TARGETARCH} AS bazel_context

ARG TARGETARCH

ENV BAZELISK_VERSION="v1.9.0"
ENV BAZELISK_BASE_URL="https://github.com/bazelbuild/bazelisk/releases/download"
ENV BAZELISK_BIN="bazelisk-linux-${TARGETARCH}"
ENV BAZELISK_URL="${BAZELISK_BASE_URL}/${BAZELISK_VERSION}/${BAZELISK_BIN}"

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    ca-certificates

RUN wget -nv ${BAZELISK_URL}
RUN chmod +x ${BAZELISK_BIN}
RUN mv ${BAZELISK_BIN} /usr/local/bin/bazel

########################
# Final image for proxy
########################

FROM ubuntu:xenial AS build_env_proxy_amd64
ENV UBUNTU_RELEASE_CODE_NAME=xenial
FROM ubuntu:bionic AS build_env_proxy_arm64
ENV UBUNTU_RELEASE_CODE_NAME=bionic
# hadolint ignore=DL3006
FROM build_env_proxy_${TARGETARCH} AS build_env_proxy

WORKDIR /

# Version from build arguments
ARG VERSION

# Labels used by Docker
LABEL "io.istio.repo"="https://github.com/istio/tools"
LABEL "io.istio.version"="${VERSION}"

# Docker
ENV DOCKER_VERSION=5:20.10.7~3-0~ubuntu-${UBUNTU_RELEASE_CODE_NAME}
ENV CONTAINERD_VERSION=1.4.6-1

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

# LLVM support
ENV LLVM_DIRECTORY=/usr/lib/llvm
ENV PATH=${LLVM_DIRECTORY}/bin:$PATH

# Avoid interactive input when installing tshark
ENV DEBIAN_FRONTEND=noninteractive

# required for binary tools: ca-certificates, gcc, libc-dev, git, iptables, libltdl7
# required for general build: make, wget, curl, ssh
# required for python: python3, pkg-config
# hadolint ignore=DL3008, DL3009
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common \
    ca-certificates \
    gcc \
    ssh \
    libltdl7 \
    libc-dev \
    make \
    pkg-config \
    python3 \
    python3-setuptools \
    daemon \
    wget \
    jq \
    tshark

# Build git from source. Golang now requires a recent git version
# hadolint ignore=DL3003,DL3009,DL4001
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    libssl-dev \
    libcurl4-gnutls-dev \
    libexpat1-dev \
    zlib1g-dev \
    gettext \
    unzip && \
    wget -q https://github.com/git/git/archive/v2.25.1.zip -O git.zip && \
    unzip git.zip && \
    cd git-* && \
    make prefix=/usr/local all && \
    make prefix=/usr/local install && \
    cd .. && rm -r git-*

# Docker including docker-ce, docker-ce-cli, and containerd.io
ADD https://download.docker.com/linux/ubuntu/gpg /tmp/docker-key
RUN apt-key add /tmp/docker-key
ARG TARGETARCH
RUN add-apt-repository "deb [arch=${TARGETARCH}] https://download.docker.com/linux/ubuntu ${UBUNTU_RELEASE_CODE_NAME} stable"
# hadolint ignore=DL3009
RUN apt-get update && apt-get -y install --no-install-recommends \
    docker-ce="${DOCKER_VERSION}" \
    docker-ce-cli="${DOCKER_VERSION}" \
    containerd.io="${CONTAINERD_VERSION}"

# Run dockerd in CI
COPY prow-entrypoint.sh /usr/local/bin/entrypoint
RUN chmod +x /usr/local/bin/entrypoint

# CMake
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# ubuntu 18.04 without arm64 version see https://apt.kitware.com/
# hadolint ignore=DL4001
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) \
            curl -fsSL https://apt.kitware.com/keys/kitware-archive-latest.asc | apt-key add - ; \
            apt-add-repository "deb https://apt.kitware.com/ubuntu/ ${UBUNTU_RELEASE_CODE_NAME} main"; \
            ;; \
        *) echo "skip" ;; \
    esac;

# binary dependencies to build envoy at v1.12.0
# https://github.com/envoyproxy/envoy/blob/v1.12.0/bazel/README.md
# hadolint ignore=DL3008,DL3009
RUN apt-get update && apt-get install -y --no-install-recommends \
    autoconf \
    automake \
    cmake \
    libtool \
    ninja-build \
    python \
    unzip \
    virtualenv

COPY --from=binary_tools_context /out/ /
COPY --from=binary_tools_context /usr/local/go /usr/local/go
COPY --from=gn_context /gn/gn /usr/local/bin/gn
COPY --from=bazel_context /usr/local/bin /usr/local/bin
COPY --from=clang_context ${LLVM_DIRECTORY}/lib ${LLVM_DIRECTORY}/lib
COPY --from=clang_context ${LLVM_DIRECTORY}/bin ${LLVM_DIRECTORY}/bin
COPY --from=clang_context ${LLVM_DIRECTORY}/include ${LLVM_DIRECTORY}/include

RUN echo "${LLVM_DIRECTORY}/lib" | tee /etc/ld.so.conf.d/llvm.conf
RUN ldconfig

COPY proxy-tsan-instrumented-libcxx.sh proxy-tsan-instrumented-libcxx.sh
RUN ./proxy-tsan-instrumented-libcxx.sh

COPY install-python-3.8.sh install-python-3.8.sh
RUN ./install-python-3.8.sh

# su-exec is used in place of complex sudo setup operations
RUN chmod u+sx /usr/bin/su-exec

COPY bashrc /home/.bashrc

# mountpoints are mandatory for any host mounts.
# mountpoints in /config are special.
RUN mkdir -p /go && \
    mkdir -p /gocache && \
    mkdir -p /gobin && \
    mkdir -p /config/.docker && \
    mkdir -p /config/.config/gcloud && \
    mkdir -p /config/.kube && \
    mkdir -p /config-copy && \
    mkdir -p /home/.cache && \
    mkdir -p /home/.helm && \
    mkdir -p /home/.gsutil

# TODO must sort out how to use uid mapping in docker so these don't need to be 777
# They are created as root 755.  As a result they are not writeable, which fails in
# the developer environment as a volume or bind mount inherits the permissions of
# the directory mounted rather then overriding with the permission of the volume file.
RUN chmod 777 /go && \
    chmod 777 /gocache && \
    chmod 777 /gobin && \
    chmod 777 /config && \
    chmod 777 /config/.docker && \
    chmod 777 /config/.config/gcloud && \
    chmod 777 /config/.kube && \
    chmod 777 /home/.cache && \
    chmod 777 /home/.helm && \
    chmod 777 /home/.gsutil

# Clean up stuff we don't need in the final image
RUN rm -rf /var/lib/apt/lists/*
RUN rm -fr /usr/share/python
RUN rm -fr /usr/share/bash-completion
RUN rm -fr /usr/share/bug
RUN rm -fr /usr/share/doc
RUN rm -fr /usr/share/dh-python
RUN rm -fr /usr/share/locale
RUN rm -fr /usr/share/man
RUN rm -fr /tmp/*

# Run config setup in local environments
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint

ENTRYPOINT ["/usr/local/bin/docker-entrypoint"]

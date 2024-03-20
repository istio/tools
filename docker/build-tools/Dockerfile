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
# The container currently supports linux/amd64 and linux/arm64 builds.
#
# The container is built using different contexts, one per execution environment (plain binaries, Ruby, nodejs), and
# then combined into the final image.
#
# We pin versions of stuff we install. Modify the various XXX_VERSION variables within the individual build contexts
# in order to control these versions.

################
# Binary tools
################
ARG GOLANG_IMAGE=golang:1.22.1-bullseye
# hadolint ignore=DL3006
FROM ${GOLANG_IMAGE} as binary_tools_context
# TARGETARCH is an automatic platform ARG enabled by Docker BuildKit.
ARG TARGETARCH

# Istio tools SHA that we use for this image
ARG ISTIO_TOOLS_SHA

# Pinned versions of stuff we pull in, keep this list sorted alphabetically
ENV BENCHSTAT_VERSION=9c9101da8316
ENV BOM_VERSION=v0.6.0
ENV BUF_VERSION=v1.30.0
ENV COSIGN_VERSION=v2.2.3
ENV CRANE_VERSION=v0.19.1
ENV GCLOUD_VERSION=467.0.0
ENV GCR_AUTH_VERSION=2.1.21
ENV GH_VERSION=2.45.0
ENV GOCOVMERGE_VERSION=b5bfa59ec0adc420475f97f89b58045c721d761c
ENV GOIMPORTS_VERSION=v0.17.0
# When updating the golangci version, you may want to update the common-files config/.golangci* files as well.
ENV GOLANGCI_LINT_VERSION=v1.56.2
ENV GOLANG_GRPC_PROTOBUF_VERSION=v1.3.0
ENV GOLANG_PROTOBUF_VERSION=v1.33.0
ENV GO_BINDATA_VERSION=v3.1.2
ENV GO_JUNIT_REPORT_VERSION=df0ed838addb0fa189c4d76ad4657f6007a5811c
ENV HADOLINT_VERSION=v2.12.0
ENV HELM3_VERSION=v3.14.3
ENV HUGO_VERSION=0.124.1
ENV JB_VERSION=v0.3.1
ENV JSONNET_VERSION=v0.20.0
ENV JUNIT_MERGER_VERSION=adf1545b49509db1f83c49d1de90bbcb235642a8
ENV K8S_CODE_GENERATOR_VERSION=1.29.2
# From Feb 15, 2024
ENV K8S_TEST_INFRA_VERSION=cf57a1d7fe5f8b7676ae7ebe9098ddbcad3d3023
ENV KIND_VERSION=v0.22.0
ENV KPT_VERSION=v1.0.0-beta.49
ENV KUBECTL_VERSION=1.29.3
ENV KUBECTX_VERSION=0.9.5
ENV KUBETEST2_VERSION=b019714a389563c9a788f119f801520d059b6533
ENV OC_VERSION=4.15.0
ENV ORAS_VERSION=1.1.0
ENV OTEL_CLI_VERSION=v0.4.1
ENV PROTOC_GEN_GRPC_GATEWAY_VERSION=v1.16.0
ENV PROTOC_VERSION=26.0
ENV PROTOLOCK_VERSION=v0.17.0
ENV SHELLCHECK_VERSION=v0.9.0
ENV SU_EXEC_VERSION=0.3.1
ENV TRIVY_VERSION=0.50.0
ENV YQ_VERSION=4.42.1

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
        x86_64) PROTOC_ZIP=protoc-${PROTOC_VERSION}-linux-x86_64.zip;; \
        aarch64) PROTOC_ZIP=protoc-${PROTOC_VERSION}-linux-aarch_64.zip;; \
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
RUN go install -ldflags="-s -w" github.com/nilslice/protolock/cmd/protolock@${PROTOLOCK_VERSION}
RUN go install -ldflags="-s -w" golang.org/x/tools/cmd/goimports@${GOIMPORTS_VERSION}
RUN go install -ldflags="-s -w" github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}
RUN go install -ldflags="-s -w" github.com/go-bindata/go-bindata/go-bindata@${GO_BINDATA_VERSION}
RUN go install -ldflags="-s -w" github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@${PROTOC_GEN_GRPC_GATEWAY_VERSION}
RUN go install -ldflags="-s -w" github.com/google/go-jsonnet/cmd/jsonnet@${JSONNET_VERSION}
RUN go install -ldflags="-s -w" github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@${JB_VERSION}
RUN go install -ldflags="-s -w" github.com/istio/go-junit-report@${GO_JUNIT_REPORT_VERSION}
RUN go install -ldflags="-s -w" sigs.k8s.io/bom/cmd/bom@${BOM_VERSION}
RUN go install -ldflags="-s -w" sigs.k8s.io/kind@${KIND_VERSION}
RUN go install -ldflags="-s -w" github.com/wadey/gocovmerge@${GOCOVMERGE_VERSION}
RUN go install -ldflags="-s -w" github.com/imsky/junit-merger/src/junit-merger@${JUNIT_MERGER_VERSION}
RUN go install -ldflags="-s -w" golang.org/x/perf/cmd/benchstat@${BENCHSTAT_VERSION}
RUN go install -ldflags="-s -w" github.com/google/go-containerregistry/cmd/crane@${CRANE_VERSION}
RUN go install -ldflags="-s -w" github.com/equinix-labs/otel-cli@${OTEL_CLI_VERSION}
RUN go install -ldflags="-s -w" github.com/mikefarah/yq/v4@v${YQ_VERSION}
# Install latest version of Istio-owned tools in this release
RUN go install -ldflags="-s -w" \
  istio.io/tools/cmd/protoc-gen-docs@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/annotations_prep@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/envvarlinter@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/testlinter@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/protoc-gen-golang-deepcopy@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/protoc-gen-golang-jsonshim@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/kubetype-gen@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/license-lint@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/gen-release-notes@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/org-gen@${ISTIO_TOOLS_SHA} \
  istio.io/tools/cmd/protoc-gen-crd@${ISTIO_TOOLS_SHA}
RUN go install -ldflags="-s -w" \
  k8s.io/code-generator/cmd/applyconfiguration-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION} \
  k8s.io/code-generator/cmd/defaulter-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION} \
  k8s.io/code-generator/cmd/client-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION} \
  k8s.io/code-generator/cmd/lister-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION} \
  k8s.io/code-generator/cmd/informer-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION} \
  k8s.io/code-generator/cmd/deepcopy-gen@kubernetes-${K8S_CODE_GENERATOR_VERSION} \
  k8s.io/code-generator/cmd/go-to-protobuf@kubernetes-${K8S_CODE_GENERATOR_VERSION}
RUN go install -ldflags="-s -w" \
    sigs.k8s.io/kubetest2@${KUBETEST2_VERSION} \
    sigs.k8s.io/kubetest2/kubetest2-gke@${KUBETEST2_VERSION} \
    sigs.k8s.io/kubetest2/kubetest2-tester-exec@${KUBETEST2_VERSION}
RUN go install -ldflags="-s -w" \
  k8s.io/test-infra/robots/pr-creator@${K8S_TEST_INFRA_VERSION} \
  k8s.io/test-infra/prow/cmd/peribolos@${K8S_TEST_INFRA_VERSION} \
  k8s.io/test-infra/pkg/benchmarkjunit@${K8S_TEST_INFRA_VERSION}

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
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) HADOLINT_BINARY=hadolint-Linux-x86_64;; \
        aarch64) HADOLINT_BINARY=hadolint-Linux-arm64;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    \
    wget -nv -O ${OUTDIR}/usr/bin/hadolint https://github.com/hadolint/hadolint/releases/download/${HADOLINT_VERSION}/${HADOLINT_BINARY}; \
    chmod 555 ${OUTDIR}/usr/bin/hadolint

# Hugo static site generator
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) HUGO_TAR=hugo_${HUGO_VERSION}_Linux-64bit.tar.gz;; \
        aarch64) HUGO_TAR=hugo_${HUGO_VERSION}_Linux-ARM64.tar.gz;; \
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
ADD https://github.com/NobodyXu/su-exec/archive/refs/tags/v${SU_EXEC_VERSION}.tar.gz /tmp
RUN tar -xzvf v${SU_EXEC_VERSION}.tar.gz
WORKDIR /tmp/su-exec-${SU_EXEC_VERSION}
# Setting LDFLAGS is needed here, upstream uses '--icf=all' which our linker  doesn't have
RUN LDFLAGS="-fvisibility=hidden -Wl,-O2 -Wl,--discard-all -Wl,--strip-all -Wl,--as-needed -Wl,--gc-sections" make
RUN cp -a su-exec ${OUTDIR}/usr/bin

ADD https://github.com/GoogleContainerTools/kpt/releases/download/${KPT_VERSION}/kpt_linux_${TARGETARCH} ${OUTDIR}/usr/bin/kpt
RUN chmod 555 ${OUTDIR}/usr/bin/kpt

# Install gcloud command line tool
# Install gcloud beta component
# Install GKE auth plugin
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) GCLOUD_TAR_FILE="google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz" ;; \
        aarch64) GCLOUD_TAR_FILE="google-cloud-sdk-${GCLOUD_VERSION}-linux-arm.tar.gz" ;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    \
    wget -nv "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/${GCLOUD_TAR_FILE}"; \
    tar -xzvf ."/${GCLOUD_TAR_FILE}" -C "${OUTDIR}/usr/local" && rm "${GCLOUD_TAR_FILE}"; \
    ${OUTDIR}/usr/local/google-cloud-sdk/bin/gcloud components install beta --quiet; \
    ${OUTDIR}/usr/local/google-cloud-sdk/bin/gcloud components install alpha --quiet; \
    ${OUTDIR}/usr/local/google-cloud-sdk/bin/gcloud components install gke-gcloud-auth-plugin --quiet; \
    rm -rf ${OUTDIR}/usr/local/google-cloud-sdk/.install/.backup \
    rm -rf ${OUTDIR}/usr/local/google-cloud-sdk/bin/anthoscli

# Install cosign (for signing build artifacts) and verify signature
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN set -eux; \
    wget -nv -O /tmp/cosign https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/cosign-linux-${TARGETARCH} \
    && wget -nv -O - https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/cosign-linux-${TARGETARCH}.sig | base64 -d > /tmp/cosign.sig \
    && wget -nv -O /tmp/cosign-pubkey https://raw.githubusercontent.com/sigstore/cosign/main/release/release-cosign.pub \
    && openssl dgst -sha256 -verify /tmp/cosign-pubkey -signature /tmp/cosign.sig /tmp/cosign \
    && chmod +x /tmp/cosign \
    && mv /tmp/cosign ${OUTDIR}/usr/bin/ || exit 1

# Install ORAS for OCI artifact pushing and pulling
RUN set -eux; \
    wget -nv "https://github.com/oras-project/oras/releases/download/v${ORAS_VERSION}/oras_${ORAS_VERSION}_linux_amd64.tar.gz" \
    && mkdir -p /tmp/oras-install/ \
    && tar -zxf oras_${ORAS_VERSION}_*.tar.gz -C /tmp/oras-install/ \
    &&  mv /tmp/oras-install/oras ${OUTDIR}/usr/bin/ \
    && rm -rf oras_${ORAS_VERSION}_*.tar.gz /tmp/oras-install/

# Trivy container scanner
RUN set -eux; \
    \
    case $(uname -m) in \
    x86_64) \
    TRVIY_DEB_NAME="trivy_${TRIVY_VERSION}_Linux-64bit.deb"; \
    ;; \
    aarch64) \
    TRVIY_DEB_NAME="trivy_${TRIVY_VERSION}_Linux-ARM64.deb"; \
    ;; \
    *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    wget -nv -O "/tmp/${TRVIY_DEB_NAME}" "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/${TRVIY_DEB_NAME}"; \
    apt-get -y install --no-install-recommends -f "/tmp/${TRVIY_DEB_NAME}"; \
    rm "/tmp/${TRVIY_DEB_NAME}"; \
    mv /usr/bin/trivy ${OUTDIR}/usr/bin/

# Install kubectx and kubens
ADD https://github.com/ahmetb/kubectx/releases/download/v${KUBECTX_VERSION}/kubectx /tmp
RUN mv /tmp/kubectx ${OUTDIR}/usr/bin/kubectx
RUN chmod 555 ${OUTDIR}/usr/bin/kubectx
ADD https://github.com/ahmetb/kubectx/releases/download/v${KUBECTX_VERSION}/kubens /tmp
RUN mv /tmp/kubens ${OUTDIR}/usr/bin/kubens
RUN chmod 555 ${OUTDIR}/usr/bin/kubens

# Install oc - OpenShift command line tool
RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) OC_BINARY=openshift-client-linux;; \
        aarch64) OC_BINARY=openshift-client-linux-arm64;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    \
    wget -nv -O /tmp/oc.tar.gz https://mirror.openshift.com/pub/openshift-v4/clients/ocp/${OC_VERSION}/${OC_BINARY}.tar.gz; \
    tar zxf /tmp/oc.tar.gz -C /tmp oc; \
    mv /tmp/oc ${OUTDIR}/usr/bin/oc; \
    chmod +x ${OUTDIR}/usr/bin/oc; \
    rm -rf /tmp/oc*

# Move Go tools to their final location
RUN mv /tmp/go/bin/* ${OUTDIR}/usr/bin

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
ENV NODEJS_VERSION=20.11.1
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
        x86_64) NODEJS_TAR=node-v${NODEJS_VERSION}-linux-x64.tar.gz;; \
        aarch64) NODEJS_TAR=node-v${NODEJS_VERSION}-linux-arm64.tar.gz;; \
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
ENV FPM_VERSION=v1.15.1
ENV HTMLPROOFER_VERSION=3.19.4
ENV LICENSEE_VERSION=9.16.0
ENV MDL_VERSION=0.12.0
ENV NOKOGIRI_VERSION=1.14.5

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
RUN gem install --no-wrappers --no-document nokogiri -v ${NOKOGIRI_VERSION}
RUN gem install --no-wrappers --no-document html-proofer -v ${HTMLPROOFER_VERSION}
RUN gem install --no-wrappers --no-document awesome_bot -v ${AWESOMEBOT_VERSION}
RUN gem install --no-wrappers --no-document licensee -v ${LICENSEE_VERSION}
# hadolint ignore=DL3003,DL3028
RUN mkdir fpm && \
    cd fpm && \
    git init && \
    git remote add origin https://github.com/jordansissel/fpm && \
    git fetch --depth 1 origin ${FPM_VERSION} && \
    git checkout FETCH_HEAD && \
    make install && \
    cd .. && rm -rf fpm

##############
# Python
##############

FROM ubuntu:jammy as python_context

ENV DEBIAN_FRONTEND=noninteractive

# Pinned versions of stuff we pull in
ENV AUTOPEP8_VERSION=2.0.4
ENV PYCODESTYLE_VERSION=2.11.1
ENV JWCRYPTO_VERSION=1.5.0
ENV PIP_INSTALL_VERSION=23.1.2
ENV PYGITHUB_VERSION=1.58.2
ENV PYTHON_PROTOBUF_VERSION=4.23.2
ENV PYYAML_VERSION=6.0
ENV REQUESTS_VERSION=2.31.0
ENV YAMLLINT_VERSION=1.32.0

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
RUN python3 -m pip install --no-cache-dir pycodestyle==${PYCODESTYLE_VERSION}
RUN python3 -m pip install --no-cache-dir --no-binary :all: autopep8==${AUTOPEP8_VERSION}
RUN python3 -m pip install --no-cache-dir yamllint==${YAMLLINT_VERSION}
RUN python3 -m pip install --no-cache-dir requests==${REQUESTS_VERSION}
RUN python3 -m pip install --no-cache-dir protobuf==${PYTHON_PROTOBUF_VERSION}
RUN python3 -m pip install --no-cache-dir PyYAML==${PYYAML_VERSION}
RUN python3 -m pip install --no-cache-dir jwcrypto==${JWCRYPTO_VERSION}
RUN python3 -m pip install --no-cache-dir PyGithub==${PYGITHUB_VERSION}

#############
# Base OS
#############

FROM ubuntu:jammy as base_os_context

ENV DEBIAN_FRONTEND=noninteractive

ENV CONTAINERD_VERSION=1.6.28-2
ENV DOCKER_VERSION=5:26.0.0-1~ubuntu.22.04~jammy
ENV DOCKER_BUILDX_VERSION=0.13.1-1~ubuntu.22.04~jammy
ENV RUST_VERSION=1.76.0

ENV OUTDIR=/out

# required for binary tools: ca-certificates, gcc, libc-dev, git, iptables, libltdl7, less
# required for general build: make, wget, curl, ssh, rpm
# required for ruby: libcurl4-openssl-dev
# required for python: python3, pkg-config
# required for ebpf build: clang,llvm,libbpf-dev
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common \
    ca-certificates \
    cmake \
    cmake-data \
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
    file \
    libclang-dev \
    iproute2 \
    ipset \
    rsync \
    clang \
    llvm \
    libbpf-dev \
    net-tools \
    ninja-build \
    sudo

# Fix Docker issue
RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy

# Docker including docker-ce, docker-ce-cli, and containerd.io
ADD https://download.docker.com/linux/ubuntu/gpg /tmp/docker-key
RUN apt-key add /tmp/docker-key
ARG TARGETARCH
RUN add-apt-repository "deb [arch=${TARGETARCH}] https://download.docker.com/linux/ubuntu $(lsb_release -sc) stable"
RUN apt-get update
RUN apt-get -y install --no-install-recommends docker-ce="${DOCKER_VERSION}" docker-ce-cli="${DOCKER_VERSION}" containerd.io="${CONTAINERD_VERSION}" docker-buildx-plugin="${DOCKER_BUILDX_VERSION}"
RUN sed -i 's/ulimit -Hn/# ulimit -Hn/g' /etc/init.d/docker

ENV CARGO_HOME=/home/.cargo
ENV RUSTUP_HOME=/home/.rustup
# hadolint ignore=DL4006
RUN curl --proto '=https' -v --tlsv1.2 -sSf https://sh.rustup.rs | \
    sh -s -- -y -v --default-toolchain ${RUST_VERSION} --profile minimal --component rustfmt clippy &&\
    /home/.cargo/bin/rustup default ${RUST_VERSION} &&\
    mv /home/.cargo/bin/* /usr/bin

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
# Go sleeps for 1s after tests with out this, see https://github.com/golang/go/issues/61852
ENV GORACE="atexit_sleep_ms=0"
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
    mkdir -p /home/.cargo/registry && \
    mkdir -p /home/.cargo/git && \
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
    chmod 777 /home/.cargo && \
    chmod 777 /home/.cargo/registry && \
    chmod 777 /home/.cargo/git && \
    chmod 777 /home/.helm && \
    chmod 777 /home/.gsutil

WORKDIR /

ENTRYPOINT ["/usr/local/bin/docker-entrypoint"]


##############
# Clang+LLVM
##############

FROM ubuntu:18.04 AS clang_context_amd64
ENV UBUNTU_RELEASE_CODE_NAME=bionic
FROM ubuntu:20.04 AS clang_context_arm64
ENV UBUNTU_RELEASE_CODE_NAME=focal
# hadolint ignore=DL3006
FROM clang_context_${TARGETARCH} AS clang_context

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends \
    xz-utils \
    wget \
    ca-certificates \
    software-properties-common \
    curl \
    ninja-build \
    gpg-agent \
    libtinfo5

ENV LLVM_VERSION=14.0.0
ENV LLVM_BASE_URL=https://github.com/llvm/llvm-project/releases/download/llvmorg-${LLVM_VERSION}
ENV LLVM_DIRECTORY=/usr/lib/llvm

RUN set -eux; \
    \
    case $(uname -m) in \
        x86_64) \
               LLVM_ARCHIVE=clang+llvm-${LLVM_VERSION}-x86_64-linux-gnu-ubuntu-18.04 \
               LLVM_ARTIFACT=clang+llvm-${LLVM_VERSION}-x86_64-linux-gnu-ubuntu-18.04;; \
        aarch64)  \
               LLVM_ARCHIVE=clang+llvm-${LLVM_VERSION}-aarch64-linux-gnu \
               LLVM_ARTIFACT=clang+llvm-${LLVM_VERSION}-aarch64-linux-gnu;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    \
    wget -nv ${LLVM_BASE_URL}/${LLVM_ARTIFACT}.tar.xz; \
    tar -xJf ${LLVM_ARTIFACT}.tar.xz -C /tmp; \
    mkdir -p ${LLVM_DIRECTORY}; \
    mv /tmp/${LLVM_ARCHIVE}/* ${LLVM_DIRECTORY}/

# CMake >=3.13.4
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
RUN apt-get update && apt-get install -y --no-install-recommends cmake

# TSan instrumented libc++.
ENV PATH=${LLVM_DIRECTORY}/bin:$PATH

WORKDIR /tmp
COPY proxy-tsan-instrumented-libcxx.sh proxy-tsan-instrumented-libcxx.sh
RUN ./proxy-tsan-instrumented-libcxx.sh

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

ENV BAZELISK_VERSION="v1.18.0"
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

FROM ubuntu:18.04 AS build_env_proxy_amd64
ENV UBUNTU_RELEASE_CODE_NAME=bionic
ENV UBUNTU_RELEASE_VERSION=18.04
FROM ubuntu:20.04 AS build_env_proxy_arm64
ENV UBUNTU_RELEASE_CODE_NAME=focal
ENV UBUNTU_RELEASE_VERSION=20.04
# hadolint ignore=DL3006
FROM build_env_proxy_${TARGETARCH} AS build_env_proxy

ENV DOCKER_BUILDX_VERSION=0.10.5-1~ubuntu.${UBUNTU_RELEASE_VERSION}~${UBUNTU_RELEASE_CODE_NAME}

WORKDIR /

# Version from build arguments
ARG VERSION

# Labels used by Docker
LABEL "io.istio.repo"="https://github.com/istio/tools"
LABEL "io.istio.version"="${VERSION}"

# Docker
ENV DOCKER_VERSION=5:24.0.2-1~ubuntu.${UBUNTU_RELEASE_VERSION}~${UBUNTU_RELEASE_CODE_NAME}
ENV CONTAINERD_VERSION=1.6.21-1

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
    rsync \
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
    wget -q https://github.com/git/git/archive/v2.39.1.zip -O git.zip && \
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
    containerd.io="${CONTAINERD_VERSION}" \
    docker-buildx-plugin="${DOCKER_BUILDX_VERSION}"

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
    libncurses5 \
    libtinfo5 \
    virtualenv

COPY --from=binary_tools_context /out/ /
COPY --from=binary_tools_context /usr/local/go /usr/local/go
COPY --from=gn_context /gn/gn /usr/local/bin/gn
COPY --from=bazel_context /usr/local/bin /usr/local/bin
COPY --from=clang_context ${LLVM_DIRECTORY}/lib ${LLVM_DIRECTORY}/lib
COPY --from=clang_context ${LLVM_DIRECTORY}/bin ${LLVM_DIRECTORY}/bin
COPY --from=clang_context ${LLVM_DIRECTORY}/include ${LLVM_DIRECTORY}/include
COPY --from=clang_context /opt/libcxx_tsan /opt/libcxx_tsan

COPY install-python.sh install-python.sh
RUN ./install-python.sh

RUN echo "${LLVM_DIRECTORY}/lib" | tee /etc/ld.so.conf.d/llvm.conf
RUN ldconfig

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
RUN rm -rf /var/lib/apt/lists/* && \
    rm -fr /usr/share/python && \
    rm -fr /usr/share/bash-completion && \
    rm -fr /usr/share/bug && \
    rm -fr /usr/share/doc && \
    rm -fr /usr/share/dh-python && \
    rm -fr /usr/share/locale  && \
    rm -fr /usr/share/man && \
    rm -fr /tmp/*

# Run config setup in local environments
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint

ENTRYPOINT ["/usr/local/bin/docker-entrypoint"]

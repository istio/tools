#!/bin/bash

set -eux

export DEBIAN_FRONTEND=noninteractive; \
     ln -fs /usr/share/zoneinfo/America/New_York /etc/localtime; \
     DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y tzdata && dpkg-reconfigure --frontend noninteractive tzdata

DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    build-essential \
    ca-certificates \
    curl \
    gnupg2 \
    software-properties-common \
    unzip \
    xz-utils \
    numactl \
    libnuma1 \
    libnuma-dev \
    libc6-dev \
    clang-9 \
    llvm-9 \
    libtinfo-dev \
    libtinfo6 \
    libtinfo5 \
    llvm-6.0-dev \
    llvm \
    wget

wget -nv https://github.com/hadolint/hadolint/archive/refs/tags/v${HADOLINT_VERSION}.tar.gz
tar xvf v${HADOLINT_VERSION}.tar.gz

wget -nv https://github.com/commercialhaskell/stack/releases/download/v2.7.1/stack-2.7.1-linux-aarch64.tar.gz; \
    tar xvf stack-2.7.1-linux-aarch64.tar.gz; \
    cp ./stack-2.7.1-linux-aarch64/stack /usr/bin

cd hadolint-${HADOLINT_VERSION}; \
    stack init --force; \
    stack setup; \
    stack install

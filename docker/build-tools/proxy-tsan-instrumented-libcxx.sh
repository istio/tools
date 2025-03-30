#!/usr/bin/env bash

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

set -eux

# gcc-11, need to build instrumented LLVM libc++ for tsan testing.
add-apt-repository -y ppa:ubuntu-toolchain-r/test
apt-get update && apt-get install -y --no-install-recommends g++-11
update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-11 1000

# Instrumented libcxx built from LLVM source, used for tsan testing.
# See envoy dev guide for more info: https://github.com/envoyproxy/envoy-build-tools
# should use same llvm version of build env
LLVM_VERSION=${LLVM_VERSION:-18.1.8}

WORKDIR=$(mktemp -d)
pushd "${WORKDIR}" || exit

wget -q -O -  "https://github.com/llvm/llvm-project/archive/llvmorg-${LLVM_VERSION}.tar.gz" | tar zx

sysctl vm.mmap_rnd_bits
cmake --version

pushd "llvm-project-llvmorg-${LLVM_VERSION}"
LIBCXX_PATH=${LIBCXX_PATH:-tsan}
cmake -GNinja \
    -B "${LIBCXX_PATH}" \
    -S "runtimes" \
    -DLLVM_ENABLE_RUNTIMES="libcxxabi;libcxx;libunwind" \
    -DLLVM_USE_LINKER=lld \
    -DLLVM_USE_SANITIZER="Thread" \
    -DCMAKE_BUILD_TYPE=RelWithDebInfo \
    -DCMAKE_C_COMPILER=clang \
    -DCMAKE_CXX_COMPILER=clang++ \
    -DCMAKE_INSTALL_PREFIX="/opt/libcxx_${LIBCXX_PATH}" \
    -DCMAKE_POSITION_INDEPENDENT_CODE=ON
ninja -C "${LIBCXX_PATH}" install-cxx install-cxxabi

rm -rf /opt/libcxx_tsan/include
popd

popd || exit

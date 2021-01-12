# gcc-9, need to build instrumented LLVM libc++ for tsan testing.
add-apt-repository -y ppa:ubuntu-toolchain-r/test
apt-get update && apt-get install -y --no-install-recommends g++-9
update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-9 1000
update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-9 1000
update-alternatives --config gcc
update-alternatives --config g++

# Instrumented libcxx built from LLVM source, used for tsan testing.
# See envoy dev guide for more info: https://github.com/envoyproxy/envoy/blob/v1.17.0/bazel/README.md#sanitizers
LLVM_VERSION=10.0.1
LLVM_ARCHIVE=llvmorg-${LLVM_VERSION}.tar.gz
LLVM_ARCHIVE_URL=https://github.com/llvm/llvm-project/archive/${LLVM_ARCHIVE}
wget ${LLVM_ARCHIVE_URL}
tar -xzf ${LLVM_ARCHIVE} -C /tmp
mkdir tsan
pushd tsan

cmake \
-GNinja \
-DLLVM_ENABLE_PROJECTS="libcxxabi;libcxx" \
-DLLVM_USE_LINKER=lld \
-DLLVM_USE_SANITIZER=Thread \
-DCMAKE_BUILD_TYPE=Release \
-DCMAKE_C_COMPILER=clang \
-DCMAKE_CXX_COMPILER=clang++ \
-DCMAKE_INSTALL_PREFIX="/opt/libcxx_tsan" "/tmp/llvm-project-llvmorg-${LLVM_VERSION}/llvm"

ninja install-cxx install-cxxabi

rm -rf /opt/libcxx_tsan/include
popd
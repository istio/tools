#!/bin/bash

# Copyright 2019 Tetrate
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

VERSION="12.0.1"

curl -sSL "https://raw.githubusercontent.com/llvm/llvm-project/llvmorg-${VERSION}/llvm/utils/release/test-release.sh" | sed 's,http://llvm.org,https://llvm.org,' > /home/build/test-release.sh
chmod +x /home/build/test-release.sh

mkdir -p "${BUILD_DIR}"
chown build:build "${BUILD_DIR}"

# for 12.0+ we must use python3. Their scripts don't seem to make this happen automatically.
VIRTUALENV_PYTHON=$(which python3)
export VIRTUALENV_PYTHON

sudo -u build scl enable devtoolset-9 \
  "/home/build/test-release.sh -release ${VERSION} -final -triple x86_64-linux-centos7 -configure-flags '-DCOMPILER_RT_BUILD_LIBFUZZER=off' -build-dir ${BUILD_DIR}"

#!/bin/bash

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

set -ex

USR_SRC="/usr/src"
KERNEL_VERSION="$(uname -r)"
CHROMEOS_RELEASE_VERSION="$(grep 'CHROMEOS_RELEASE_VERSION' /etc/lsb-release.host | cut -d '=' -f 2)"

build_kernel()
{
  # Build the headers
  cd "${WORKING_DIR}"
  zcat /proc/config.gz > .config
  make ARCH=x86 oldconfig > /dev/null
  make ARCH=x86 prepare > /dev/null

  # Build perf
  cd tools/perf/
  make ARCH=x86  > /dev/null
  mv perf /usr/sbin/
}

prepare_node()
{
  WORKING_DIR="/linux-lakitu-${CHROMEOS_RELEASE_VERSION}"
  SOURCES_DIR="${USR_SRC}/linux-lakitu-${CHROMEOS_RELEASE_VERSION}"
  mkdir -p "${WORKING_DIR}"
  curl -s "https://storage.googleapis.com/cos-tools/${CHROMEOS_RELEASE_VERSION}/kernel-src.tar.gz" \
    | tar -xzf - -C "${WORKING_DIR}"
  build_kernel
  rm -rf "${USR_SRC}${WORKING_DIR}"
  mv "${WORKING_DIR}" "${USR_SRC}"
}

prepare_node
mkdir -p "/lib/modules/${KERNEL_VERSION}"
ln -sf "${SOURCES_DIR}" "/lib/modules/${KERNEL_VERSION}/source"
ln -sf "${SOURCES_DIR}" "/lib/modules/${KERNEL_VERSION}/build"

sysctl kernel.perf_event_paranoid=-1
sysctl kernel.kptr_restrict=0

# fire up the node exporter process, listening at the passed in address:port
node_exporter --web.listen-address "$1"


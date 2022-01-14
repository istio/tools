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

yum install -y centos-release-scl epel-release
yum update -y
yum install -y devtoolset-9-gcc devtoolset-9-gcc-c++ devtoolset-9-libatomic-devel wget unzip which make cmake3 patch subversion ncurses-devel zlib-devel \
  python-virtualenv chrpath file perl-Data-Dumper tcl python2-psutil sudo rsync

# For LLVM to pick right libstdc++
mkdir -p /usr/lib/gcc/x86_64-redhat-linux
ln -s /opt/rh/devtoolset-9/root/usr/lib/gcc/x86_64-redhat-linux/9 /usr/lib/gcc/x86_64-redhat-linux

ln -s /usr/bin/cmake3 /usr/bin/cmake

useradd build

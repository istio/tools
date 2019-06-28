#!/bin/bash

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

set -e

inplace="${1}"
SCRIPTPATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR=$(dirname "${SCRIPTPATH}")
cd "${ROOTDIR}"

pip install --upgrade autopep8

status=0
python_files_changed=$(git diff --name-only | grep -E '\.py$' || true)
if [ -n "$python_files_changed" ]; then
    options="--aggressive --aggressive"
    output=$(autopep8 -d ${options} "${python_files_changed[@]}")
    status=$?

    if [ -n "$output" ]; then
        if [ ${inplace} == "true" ]; then
          autopep8 -i ${options} "${python_files_changed[@]}"
          echo "=== autopep8 updated some files ==="
          exit 0
        else
          echo "=== autopep8 check fail ==="
          exit 1
        fi
    fi
fi
echo "== autopep8 check succeed ==="
exit 0
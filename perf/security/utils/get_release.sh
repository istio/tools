#!/bin/bash

# Copyright 2019 Istio Authors
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

release_url=""
function get_release_url() {
  local release_type=$1
  local release=$2
  case "${OSTYPE}" in
    darwin*)
      if [[ "$release_type" == "daily" ]]; then
          release_url="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${release}/istio-${release}-osx.tar.gz"
      elif [[ "$release_type" == "release" ]]; then
          release_url="https://github.com/istio/istio/releases/download/${release}/istio-${release}-osx.tar.gz"
      elif [[ "$release_type" == "pre-release" ]]; then
          release_url="https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/istio-${release}-osx.tar.gz"
      else
          release_url="Please specify RELEASETYPE"
      fi ;;
    linux*)
      if [[ "$release_type" == "daily" ]]; then
          release_url="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${release}/istio-${release}-linux.tar.gz"
      elif [[ "$release_type" == "release" ]]; then
          release_url="https://github.com/istio/istio/releases/download/${release}/istio-${release}-linux.tar.gz"
      elif [[ "$release_type" == "pre-release" ]]; then
          release_url="https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/istio-${release}-linux.tar.gz"
      else
          release_url="Please specify RELEASETYPE"
      fi ;;
    *) release_url="unsupported OS: ${OSTYPE}" ;;
  esac
}

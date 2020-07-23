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
  ostype=""
  case "${OSTYPE}" in
    darwin*) ostype="osx" ;;
    linux*) ostype="linux-amd64" ;;
    *)
        echo "unsupported OS: ${OSTYPE}"
        return ;;
  esac

  case "${release_type}" in
    daily*)
        release_url="https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/${release}/istio-${release}-${ostype}.tar.gz" ;;
    release*)
        release_url="https://github.com/istio/istio/releases/download/${release}/istio-${release}-${ostype}.tar.gz" ;;
    pre-release*)
        release_url="https://gcsweb.istio.io/gcs/istio-prerelease/prerelease/${release}/istio-${release}-${ostype}.tar.gz" ;;
    *)
        echo "Please specify RELEASETYPE"
        return ;;
  esac
  echo "Release URL is $release_url"
}

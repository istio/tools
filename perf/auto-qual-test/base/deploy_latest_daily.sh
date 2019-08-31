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

WD=$(dirname $0)
WD=$(cd $WD; pwd)

git clone https://github.com/istio/tools
cd tools/perf/istio-install
helm init --client-only
RELEASE_MAJOR_MINOR="release-1.1"
if [[ ! -z "${TARGET_VERSION}" ]];then
  RELEASE_MAJOR_MINOR="${TARGET_VERSION}"
fi
LATEST_BUILD=$(curl -L https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/$RELEASE_MAJOR_MINOR-latest.txt)
SKIP_PROMETHEUS=true HELMREPO_URL=https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/$LATEST_BUILD/charts/ RELEASE_URL=https://gcsweb.istio.io/gcs/istio-prerelease/daily-build/$LATEST_BUILD/istio-$LATEST_BUILD-linux.tar.gz ./setup_istio.sh $RELEASE_MAJOR_MINOR

# trigger redeploy on services to get new sidecars
/etc/scripts/redeploy.sh ALL

#!/bin/bash

# Copyright 2018 Istio Authors

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

# This is currently triggered by https://github.com/istio/test-infra for release qualification.

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")

# Set up inputs needed by /istio/istio/tests/upgrade/test_crossgrade.sh
# These environment variables are passed by /istio/test-infra/prow/cluster/jobs istio periodic upgrade jobs
export SOURCE_HUB=${SOURCE_HUB:-"gcr.io/istio-testing"}
export TARGET_HUB=${TARGET_HUB:-"gcr.io/istio-testing"}
export SOURCE_TAG=${SOURCE_TAG:-"1.7-dev"}
export TARGET_TAG=${TARGET_TAG:-"master"}
export SOURCE_RELEASE_PATH=${SOURCE_RELEASE_PATH:-"https://storage.googleapis.com/istio-build/dev"}
export TARGET_RELEASE_PATH=${TAGET_RELEASE_PATH:-"https://storage.googleapis.com/istio-build/dev"}
export INSTALL_OPTIONS=${INSTALL_OPTIONS:-"istioctl"}
export FROM_PATH=${FROM_PATH:-"$(mktemp -d from_dir.XXXXXX)"}
export TO_PATH=${TO_PATH:-"$(mktemp -d to_dir.XXXXXX)"}
export SOURCE_LINUX_TAR_SUFFIX=${SOURCE_LINUX_TAR_SUFFIX:-"linux-amd64.tar.gz"}
export TARGET_LINUX_TAR_SUFFIX=${TARGET_LINUX_TAR_SUFFIX:-"linux-amd64.tar.gz"}
# TEST_SCENARIO has a default value of "upgrade-downgrade", which corresponds to
# the default test scenario of upgrade followed by downgrade.
# TEST_SCENARIO can also be configured to "upgrade" or "downgrade", which corresponds to
# the upgrade-only test scenario and the downgrade-only test scenario, respectively.
# When the test scenario is "downgrade",SOURCE_HUB and SOURCE_TAG specify the version
# to downgrade from whereas TARGET_HUB and TARGET_TAG specify the version
# to downgrade to.
export TEST_SCENARIO=${TEST_SCENARIO:-"upgrade-downgrade"}

function get_git_sha() {
  local url_path=${1}
  local tag=${2}

  GIT_SHA=
  LINUX_TAR_SUFFIX=

  if [[ "${tag}" =~ "latest" ]];then
    release_version=$(echo "${tag}" | cut -d'_' -f1)
    # shellcheck disable=SC2072
    if [[ ${release_version} < "1.6" ]]; then
       LINUX_TAR_SUFFIX="linux.tar.gz"
    fi
    GIT_SHA=$(curl "${url_path}/${release_version}-dev")
  elif [[ "${tag}" == "master" ]];then
    GIT_SHA=$(curl "${url_path}/latest")
  fi
}

get_git_sha "${SOURCE_RELEASE_PATH}" "${SOURCE_TAG}"
if [[ -n "${GIT_SHA}" ]]; then
  SOURCE_TAG="${GIT_SHA}"
fi
if [[ -n "${LINUX_TAR_SUFFIX}" ]]; then
  SOURCE_LINUX_TAR_SUFFIX="${LINUX_TAR_SUFFIX}"
fi

get_git_sha "${TARGET_RELEASE_PATH}" "${TARGET_TAG}"
if [[ -n "${GIT_SHA}" ]]; then
  TARGET_TAG="${GIT_SHA}"
fi
if [[ -n "${LINUX_TAR_SUFFIX}" ]]; then
  TARGET_LINUX_TAR_SUFFIX="${LINUX_TAR_SUFFIX}"
fi

# Download and unpack istio release artifacts.
function download_untar_istio_release() {
  local url_path=${1}
  local tag=${2}
  local suffix=${3}
  local dir=${4:-.}

  # Download artifacts
  LINUX_DIST_URL="${url_path}/${tag}/istio-${tag}-${suffix}"
  echo "Downloading ${LINUX_DIST_URL}"
  wget -q "${LINUX_DIST_URL}" -P "${dir}"
  tar -xzf "${dir}/istio-${tag}-${suffix}" -C "${dir}"
}

# shellcheck disable=SC1090
source "${ROOT}/bin/setup_cluster.sh"
# Set to any non-empty value to use kubectl configured cluster instead of mason provisioned cluster.
UPGRADE_TEST_LOCAL="${UPGRADE_TEST_LOCAL:-""}"

echo "Testing upgrade and downgrade between ${SOURCE_HUB}:${SOURCE_TAG} and ${TARGET_HUB}:${TARGET_TAG}"

# Download release artifacts.
download_untar_istio_release "${SOURCE_RELEASE_PATH}" "${SOURCE_TAG}" "${SOURCE_LINUX_TAR_SUFFIX}" "${FROM_PATH}"
download_untar_istio_release "${TARGET_RELEASE_PATH}" "${TARGET_TAG}" "${TARGET_LINUX_TAR_SUFFIX}" "${TO_PATH}"

# Check https://github.com/istio/test-infra/blob/master/boskos/resources.yaml
# for existing resources types
if [[ "${UPGRADE_TEST_LOCAL}" = "" ]]; then
    export RESOURCE_TYPE="${RESOURCE_TYPE:-gke-e2e-test}"
    export OWNER='upgrade-tests'
    export USE_MASON_RESOURCE="${USE_MASON_RESOURCE:-True}"
    export CLEAN_CLUSTERS="${CLEAN_CLUSTERS:-True}"

    setup_e2e_cluster
else
    echo "Running against cluster that kubectl is configured for."
fi

# Install fortio which is needed by the upgrade test.
go get fortio.org/fortio

# Kick off tests
"${WD}/test_crossgrade.sh" \
  --from_hub="${SOURCE_HUB}" --from_tag="${SOURCE_TAG}" --from_path="${FROM_PATH}/istio-${SOURCE_TAG}" \
  --to_hub="${TARGET_HUB}" --to_tag="${TARGET_TAG}" --to_path="${TO_PATH}/istio-${TARGET_TAG}" \
  --install_options="${INSTALL_OPTIONS}" --cloud="GKE"



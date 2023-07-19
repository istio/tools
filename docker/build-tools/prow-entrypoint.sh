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

# shellcheck disable=SC2317  # Don't warn about unreachable commands in this file, they are all false positives due to tracing

TEST_START="$(date -u +%s.%N)"
GCP_REGISTRIES=${GCP_REGISTRIES:-"gcr.io,us-docker.pkg.dev"}

function read_gcp_secrets() {
  # Prevent calling with -x set
  if [[ $- = *x* ]]; then
    echo "Execution tracing must be disabled to call read_gcp_secrets"
    exit 1
  fi
  # This should be impossible, but if in case the above check is wrong turn of tracing again, just in cvase...
  { set +x; } 2>/dev/null
  readarray -t secrets < <(<<<"${GCP_SECRETS}" jq -c '.[]')
  for item in "${secrets[@]}"; do
    proj="$(<<<"$item" jq .project -r)"
    secret="$(<<<"$item" jq .secret -r)"
    env="$(<<<"$item" jq -r '.env // ""')"
    file="$(<<<"$item" jq -r '.file // ""')"
    echo "Fetching secret '${secret}' in project '${proj}'"
    value="$(gcloud secrets versions access latest --secret "${secret}" --project "${proj}")"
    if [[ "${env}" != "" ]]; then
      export "$env=$value"
    fi
    if [[ "${file}" != "" ]]; then
      mkdir -p "$(dirname "${file}")"
      echo "$value" > "${file}"
    fi
  done
}

read_gcp_secrets

# Output a message, with a timestamp matching istio log format
function log() {
  # Disable execution tracing to avoid noise
  { [[ $- = *x* ]] && was_execution_trace=1 || was_execution_trace=0; } 2>/dev/null
  { set +x; } 2>/dev/null
  delta=$(date +%s.%N --date="$TEST_START seconds ago")
  echo -e "$(date -u '+%Y-%m-%dT%H:%M:%S.%NZ')\t${delta%.*}s\t$*"
  if [[ $was_execution_trace == 1 ]]; then
    { set -x; } 2>/dev/null
  fi
}

function tracing::run() {
  # Setup a default implementation that just logs May be overridden if the repo has tracing support.
  log "Running ${1}"
  "${@:2}"
  log "Completed ${1}"
}
if [[ -f common/scripts/tracing.sh ]]; then
  # shellcheck source=/dev/null
  source "common/scripts/tracing.sh"
  tracing::extract_prow_trace
fi

set -x

log "Starting test..."

# Always enable IPv6; all tests should function with it enabled so no need to be selective.
sysctl net.ipv6.conf.all.forwarding=1
sysctl net.ipv6.conf.all.disable_ipv6=0
log "Done enabling IPv6 in Docker config."
function run_docker() {
  # Set ENABLE_DOCKER to what they specify, or otherwise enable if /var/lib/docker is enabled (required for docker)
  [[ -d "/var/lib/docker" ]] && HAS_DOCKER="true"
  ENABLE_DOCKER="${ENABLE_DOCKER:-"${HAS_DOCKER}"}"

  if [[ "${ENABLE_DOCKER}" == "true" ]]; then
    log "Enabling docker..."
    # Enable debug logs for docker daemon, and set the MTU to the external NIC MTU
    # Docker will always use 1500 as the MTU; in environments where the host has <1500 as the MTU
    # this may cause connectivity issues.
    mkdir -p /etc/docker
    primaryInterface="$(awk '$2 == 00000000 { print $1 }' /proc/net/route)"
    hostMTU="$(cat "/sys/class/net/${primaryInterface}/mtu")"
    echo "{\"debug\":true, \"mtu\":${hostMTU:-1500}}" > /etc/docker/daemon.json

    # Start docker daemon and wait for dockerd to start
    service docker start

    log "Waiting for dockerd to start..."
    while :
    do
      log "Checking for running docker daemon."
      if docker ps -q > /dev/null 2>&1; then
        log "The docker daemon is running."
        break
      fi
      sleep 1
    done
  fi
  function cleanup() {
    if [[ "${ENABLE_DOCKER}" == "true" ]]; then
      log "Starting cleanup..."
      # Cleanup all docker artifacts
      docker ps -q | xargs -r docker kill
      docker system prune -af || true
      log "Cleanup complete"
    fi
  }

  trap cleanup EXIT

  # Always try to authenticate to GCR and AR.
  gcloud auth configure-docker "${GCP_REGISTRIES}" -q || true
}

tracing::run "docker" run_docker

# Authenticate gcloud, allow failures. TODO: cleanup? We should be using workload identity everywhere
if [[ -n "${GOOGLE_APPLICATION_CREDENTIALS:-}" ]]; then
  # Jobs that need this will fail later and jobs that don't should fail because of this
  gcloud auth activate-service-account --key-file="${GOOGLE_APPLICATION_CREDENTIALS}" || true
  log "WARNING: using insecure Service Account key authentication"
fi

set +x
tracing::run "test" "$@"
EXIT_VALUE=$?
set -x

# We cleanup in the trap as well, but just in case try to clean up here as well
# shellcheck disable=SC2046
tracing::run "cleanup" cleanup

exit "${EXIT_VALUE}"

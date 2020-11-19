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

set -x

# Output a message, with a timestamp matching istio log format
function log() {
  echo -e "$(date -u '+%Y-%m-%dT%H:%M:%S.%NZ')\t$*"
}

# optionally enable ipv6 docker
export DOCKER_IN_DOCKER_IPV6_ENABLED=${DOCKER_IN_DOCKER_IPV6_ENABLED:-false}
if [[ "${DOCKER_IN_DOCKER_IPV6_ENABLED}" == "true" ]]; then
  # enable ipv6
  sysctl net.ipv6.conf.all.disable_ipv6=0
  sysctl net.ipv6.conf.all.forwarding=1
  log "Done enabling IPv6 in Docker config."
fi

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

function cleanup() {
  log "Starting cleanup..."
  # Cleanup all docker artifacts
  # shellcheck disable=SC2046
  docker kill $(docker ps -q) || true
  docker system prune -af || true
  log "Cleanup complete"
}

trap cleanup EXIT

# Authenticate gcloud, allow failures
if [[ -n "${GOOGLE_APPLICATION_CREDENTIALS:-}" ]]; then
  # Jobs that need this will fail later and jobs that don't should fail because of this
  gcloud auth activate-service-account --key-file="${GOOGLE_APPLICATION_CREDENTIALS}" || true
  gcloud auth configure-docker -q || true
fi

set +x
"$@"
EXIT_VALUE=$?
set -x

# We cleanup in the trap as well, but just in case try to clean up here as well
# shellcheck disable=SC2046
cleanup

exit "${EXIT_VALUE}"

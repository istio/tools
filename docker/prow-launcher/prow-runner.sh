#!/bin/bash

set -eux

service docker start

[[ -n ${GOPATH:-} ]] && export PATH=${GOPATH}/bin:${PATH}

# Authenticate gcloud, allow failures
if [[ -n "${GOOGLE_APPLICATION_CREDENTIALS:-}" ]]; then
  gcloud auth activate-service-account --key-file="${GOOGLE_APPLICATION_CREDENTIALS}" || true
fi

exec "$@"

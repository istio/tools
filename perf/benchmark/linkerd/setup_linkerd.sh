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

# Linkerd install script
set -ex

log() { echo "$1" >&2; }
fail() { log "$1"; exit 1; }

release="${1:-stable-2.6.0}"

log "Installing Linkerd version $release"

# To install Linkerd CLI
curl -sL https://run.linkerd.io/install | sh
# Add Linkerd to your path
export PATH=$PATH:$HOME/.linkerd2/bin
# Verify the CLI is installed and running correctly
linkerd version

# shellcheck disable=SC2046
# shellcheck disable=SC2086
kubectl create clusterrolebinding cluster-admin-binding-$USER \
    --clusterrole=cluster-admin --user=$(gcloud config get-value account)

# To check that your cluster is configured correctly and ready to install the control plane
linkerd check --pre

# shellcheck disable=SC2086
linkerd install | kubectl apply -f -

# Validate the intallation
linkerd check
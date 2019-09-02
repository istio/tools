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

release="${1:-stable-2.3.2}"

log "Installing Linkerd version $release"

curl -sL https://run.linkerd.io/install | sh
export PATH=$PATH:$HOME/.linkerd2/bin
linkerd version

kubectl create clusterrolebinding cluster-admin-binding-"$USER" \
    --clusterrole=cluster-admin --user="$(gcloud config get-value account)"

linkerd check --pre

linkerd install "--linkerd-version=$release" --proxy-auto-inject | kubectl apply -f -

linkerd check
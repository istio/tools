#!/bin/bash
# Linkerd install script 
set -ex

log() { echo "$1" >&2; }
fail() { log "$1"; exit 1; }

release="${1:-stable-2.3.2}"

log "Installing Linkerd version $release"  

curl -sL https://run.linkerd.io/install | sh
export PATH=$PATH:$HOME/.linkerd2/bin
linkerd version

kubectl create clusterrolebinding cluster-admin-binding-$USER \
    --clusterrole=cluster-admin --user=$(gcloud config get-value account)

linkerd check --pre

linkerd install --linkerd-version=$release --proxy-auto-inject | kubectl apply -f -

linkerd check 
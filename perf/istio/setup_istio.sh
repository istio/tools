#!/bin/bash
set -ex

source common.sh

DNS_DOMAIN=${DNS_DOMAIN:?"DNS_DOMAIN like v104.qualistio.org"}

WD=$(dirname $0)
WD=$(cd $WD; pwd)
mkdir -p "${WD}/tmp"

release="${1:?"release"}"
shift

setup_admin_binding
install_istio "${WD}/tmp" "${release}" $*
install_gateways


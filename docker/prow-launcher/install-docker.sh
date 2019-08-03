#!/bin/bash

set -eux

#Version: 18.06.1~ce-0~ubuntu-xenial
DOCKER_VERSION=18.06.1
VERSION_SUFFIX="~ce~3-0~ubuntu"

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
 $(lsb_release -cs) stable"
apt-get update
apt-get -qqy install docker-ce="${DOCKER_VERSION}${VERSION_SUFFIX}"

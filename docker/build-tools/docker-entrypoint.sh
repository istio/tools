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

# Copy credentials from mountpoints using su-exec
uid=$(id -u)
gid=$(id -g)

shopt -s dotglob

# Make a copy of the hosts's config secrets
su-exec 0:0 cp -aR /config/* /config-copy/

# Set the ownershp of the host's config secrets to that of the ontainer
su-exec 0:0 chown -R "${uid}":"${gid}" /config-copy

# Permit only the UID:GID to read the copy of the host's config secrets
chmod -R 700 /config-copy

# Set ownership of /home to UID:GID
su-exec 0:0 chown "${uid}":"${gid}" /home

# Copy the config secrets without chaning permissions nor ownership for
# consumption by tooolchains
cp -aR /config-copy/* /home/

exec "$@"

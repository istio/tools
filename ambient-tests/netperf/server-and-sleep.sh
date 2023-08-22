#! /bin/bash

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

set -eux
# Couldn't find a way to run netserver in the foreground.
netserver "$@"
echo "Started echo server."
ncat -e /bin/cat -k -l 6789 &
echo "Started netserver."
python ./tcp_ping/server.py &
echo "Started TCP ping server"
echo "Sleeping"
sleep 365d

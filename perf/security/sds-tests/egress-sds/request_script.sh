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

# shellcheck disable=SC2086
sleep 30
num_curl=0
num_succeed=0
while true; do
  resp_code=$(curl -s  -o /dev/null -w "%{http_code}\n" URL_TO_REPLACE)
  if [ ${resp_code} = 200 ]; then
    num_succeed=$((num_succeed+1))
  else
    echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") curl to URL_TO_REPLACE failed, response code $resp_code"
  fi
  num_curl=$((num_curl+1))
  echo "$(date +"%Y-%m-%d %H:%M:%S:%3N") Out of ${num_curl} curl, ${num_succeed} succeeded."
  sleep .5
done

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

set -ex

NUM=${NUM:?"specify the number of gateway"}
CLUSTER=${CLUSTER:?"specify the cluster for running the test"}

kubectl -n istio-system delete secret ingress-root
for s in $(kubectl -n istio-system get secrets -oname | grep "httpbin-credential*")
do
  kubectl -n istio-system delete $s
done
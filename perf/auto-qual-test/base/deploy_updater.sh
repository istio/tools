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

# this will look much better once we have kubectl 1.14 with kustomize support
kubectl create configmap qual-test-deployer --from-file=deploy_latest_daily.sh,../../bin/redeploy.sh --dry-run -o yaml | kubectl apply -f -
kubectl apply -f .
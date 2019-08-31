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

# This script spins up the standard 20 services per namespace test for as many namespaces
# as desired.
WD=$(dirname $0)
WD=$(cd "${WD}"; pwd)
cd "${WD}"

set -ex

source common.sh

NUM=${1:?"number of namespaces. 20 x this number"}
START=${2:-"0"}

# service-graph04 svc04
CMD=""
if [[ ! -z "${ECHO}" ]];then
  CMD="echo"
fi

start_servicegraphs "${NUM}" "${START}"

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

# Create csv from key-value pairs.

set -eu
shopt -s extglob
# shellcheck disable=SC1091
source scripts/config

# non csv files
for file in $RESULTS/{TCP_STREAM,TCP_CRR,TCP_RR}
do
    echo "$file"
    base=$(basename "$file")
    python ./scripts/results_to_csv.py \
        "$TEST_RUN_SEPARATOR"          \
        < "$file"                      \
        > "$RESULTS/$base.csv"
done

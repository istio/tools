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

import csv
from typing import Set, List, Dict
import sys

TEST_RUN_SEPARATOR = sys.argv[1]

fieldnames: Set[str] = set()
rows: List[Dict[str, str]] = []
row: Dict[str, str] = dict()

for line in sys.stdin:
    line = line.strip()
    if line.strip() == TEST_RUN_SEPARATOR:
        fieldnames.update(row.keys())
        rows.append(row)
        row = dict()
        continue

    line = line.split("=")
    if len(line) != 2:
        continue

    row[line[0]] = line[1]

writer = csv.DictWriter(sys.stdout, fieldnames=fieldnames)

writer.writeheader()
for row in rows:
    writer.writerow(row)

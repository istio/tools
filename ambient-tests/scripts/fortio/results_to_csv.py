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
import dotenv
import os
import json
import sys

dotenv.load_dotenv("./scripts/config.sh")

TEST_RUN_SEPARATOR = os.environ["TEST_RUN_SEPARATOR"]

# Format is
# <source mesh>:<dest mesh>
# <JSON output>
# TEST_RUN_SEPARATOR

raw = sys.stdin.read()
raw = raw.replace("\r", "").strip()
runs = raw.split(TEST_RUN_SEPARATOR)
if not runs[-1]:
    runs = runs[:-1]

writer = csv.DictWriter(
    sys.stdout,
    fieldnames=[
        "NAMESPACES",
        "THROUGHPUT",
        "THROUGHPUT_UNITS",
        "P50_LATENCY",
        "P90_LATENCY",
        "P99_LATENCY",
        "SOCKET_COUNT",
    ],
    extrasaction="ignore",
)

writer.writeheader()

for run in runs:
    row = dict()
    run = run.strip()
    run = run.split("\n")

    # Hardcode field mappings
    row["NAMESPACES"] = run[0]
    run = "".join(run[1:])
    run = json.loads(run)
    row["THROUGHPUT"] = run["ActualQPS"]
    row["THROUGHPUT_UNITS"] = "Trans/s"
    for p in run["DurationHistogram"]["Percentiles"]:
        # cursed fstring
        row[f"P{p['Percentile']}_LATENCY"] = (
            p["Value"] * 10**6
        )  # seconds to microseconds
    row["SOCKET_COUNT"] = run["SocketCount"]

    writer.writerow(row)

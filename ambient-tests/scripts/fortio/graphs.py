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

from scripts.lib import latency_graph
import os
import sys

import dotenv
import matplotlib.pyplot as plt
import pandas as pd

sys.path.append(".")  # kinda hacky, but easy


# load data
# TCP_STREAM figure
dotenv.load_dotenv("./scripts/config.sh")
FORTIO_RESULTS = os.environ["FORTIO_RESULTS"]
FORTIO_GRAPHS = os.environ["FORTIO_GRAPHS"]


if __name__ == "__main__":
    os.makedirs(FORTIO_GRAPHS, exist_ok=True)
    serial = latency_graph(f"{FORTIO_RESULTS}/serial.csv", "HTTP Echo (serial)")
    serial.savefig(f"{FORTIO_GRAPHS}/serial.png")

    serial = latency_graph(f"{FORTIO_RESULTS}/parallel.csv", "HTTP Echo (parallel)")
    serial.savefig(f"{FORTIO_GRAPHS}/parallel.png")

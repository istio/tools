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
NETPERF_RESULTS = os.environ["NETPERF_RESULTS"]
NETPERF_GRAPHS = os.environ["NETPERF_GRAPHS"]


def tcp_stream_graph():
    stream_df = pd.read_csv(
        f"./{NETPERF_RESULTS}/TCP_STREAM.csv", usecols=["THROUGHPUT", "NAMESPACES"]
    )
    gb = stream_df.groupby("NAMESPACES")
    groups = sorted(list(gb.groups.keys()))  # list for consistent ordering
    fig: plt.Figure
    ax: plt.Axes
    fig, ax = plt.subplots()

    height = [gb["THROUGHPUT"].mean()[g] for g in groups]
    yerr = [gb["THROUGHPUT"].std()[g] * 2 for g in groups]
    x = list(range(len(groups)))

    ax.errorbar(x=x, y=height, yerr=yerr, fmt="|", color="r", label=r"2 $\cdot$ stddev")
    ax.bar(x=x, height=height)

    ax.set_ylabel("Mean throughput ($10^6$ bits/second)")
    ax.set_title("TCP THROUGHPUT")
    ax.set_xticks(x, groups, rotation=11)
    ax.legend()

    fig.savefig(f"./{NETPERF_GRAPHS}/TCP_STREAM.png")


def tcp_rr_graph():
    fig = latency_graph(f"./{NETPERF_RESULTS}/TCP_RR.csv", "TCP REQUEST RESPONSE")
    fig.savefig(f"./{NETPERF_GRAPHS}/TCP_RR.png")


def tcp_crr_graph():
    fig = latency_graph(
        f"./{NETPERF_RESULTS}/TCP_RR.csv", "TCP CONNECT REQUEST RESPONSE"
    )
    fig.savefig(f"./{NETPERF_GRAPHS}/TCP_CRR.png")


if __name__ == "__main__":
    os.makedirs(NETPERF_GRAPHS, exist_ok=True)
    tcp_stream_graph()
    tcp_rr_graph()
    tcp_crr_graph()

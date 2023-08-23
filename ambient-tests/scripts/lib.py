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

import pandas as pd
import matplotlib.pyplot as plt


def latency_graph(source_csv: str, title: str):
    stream_df = pd.read_csv(
        source_csv,
        usecols=[
            "P90_LATENCY",
            "P99_LATENCY",
            "P50_LATENCY",
            "NAMESPACES",
        ],
    )
    gb = stream_df.groupby("NAMESPACES")
    groups = sorted(gb.groups.keys())  # list for consistent ordering
    fig: plt.Figure
    ax: plt.Axes
    fig, ax = plt.subplots()

    height50 = [gb["P50_LATENCY"].median()[g] for g in groups]
    height90 = [gb["P90_LATENCY"].median()[g] for g in groups]
    height99 = [gb["P99_LATENCY"].median()[g] for g in groups]
    x = list(range(len(groups)))

    ax.bar(x=x, height=height99, label="P99")
    ax.bar(x=x, height=height90, label="P90")
    ax.bar(x=x, height=height50, label="P50")

    ax.set_title(title)
    ax.set_xlabel("src:dst")
    ax.set_ylabel("Transaction speed (usec/transaction)")
    ax.legend()
    ax.set_xticks(x, groups, rotation=11)

    return fig

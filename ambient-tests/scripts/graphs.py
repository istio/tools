import os
import dotenv

import matplotlib.pyplot as plt
import pandas as pd

# load data
# TCP_STREAM figure
dotenv.load_dotenv("./scripts/config.sh")
RESULTS = os.environ["RESULTS"]
GRAPHS = os.environ["GRAPHS"]


def tcp_stream_graph():
    stream_df = pd.read_csv(
        f"./{RESULTS}/TCP_STREAM.csv", usecols=["THROUGHPUT", "NAMESPACES"]
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

    fig.savefig(f"./{GRAPHS}/TCP_STREAM.png")


def tcp_rr_graph():
    stream_df = pd.read_csv(
        f"./{RESULTS}/TCP_RR.csv",
        usecols=[
            "MAX_LATENCY",
            "P90_LATENCY",
            "P99_LATENCY",
            "P50_LATENCY",
            "STDDEV_LATENCY",
            "NAMESPACES",
        ],
    )
    gb = stream_df.groupby("NAMESPACES")
    groups = list(gb.groups.keys())  # list for consistent ordering
    groups.sort()
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

    ax.set_title("TCP REQUEST RESPONSE")
    ax.set_xlabel("src:dst")
    ax.set_ylabel("Transaction speed (usec/transaction)")
    ax.legend()
    ax.set_xticks(x, groups, rotation=11)

    fig.savefig(f"./{GRAPHS}/TCP_RR.png")


def tcp_crr_graph():
    stream_df = pd.read_csv(
        f"./{RESULTS}/TCP_CRR.csv",
        usecols=[
            "P90_LATENCY",
            "P99_LATENCY",
            "P50_LATENCY",
            "STDDEV_LATENCY",
            "NAMESPACES",
        ],
    )
    gb = stream_df.groupby("NAMESPACES")
    groups = list(gb.groups.keys())  # list for consistent ordering
    groups.sort()
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

    ax.set_title("TCP CONNECT REQUEST RESPONSE")
    ax.set_xlabel("(src:dst)")
    ax.set_ylabel("Transaction speed (usec/transaction)")
    ax.legend()
    ax.set_xticks(x, groups, rotation=11)

    fig.savefig(f"./{GRAPHS}/TCP_CRR.png")


if __name__ == "__main__":
    os.makedirs(GRAPHS, exist_ok=True)
    tcp_stream_graph()
    tcp_rr_graph()
    tcp_crr_graph()

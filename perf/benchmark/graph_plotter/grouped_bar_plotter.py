import matplotlib.pyplot as plt
import numpy as np
import pandas as pd

df = pd.read_csv('grpcbench.csv')


def build_chart(column, xlabel, ylabel, name):
    columns = ['2', '4', '8', '16', '32', '64']
    labels = [
        'baseline',
        'proxyless_to_proxyless_plaintext',
        'proxyless_to_proxyless_mtls',
        'envoy_envoy_plaintext',
        'envoy_envoy_mtls',
    ]

    x = np.arange(len(columns))
    fig, ax = plt.subplots()
    groups = []
    i = 0
    for l in labels:
        q = 'Labels=="%s"' % l
        # TODO sort by NumThreads column; this relies on data in sheet being ordered
        group_data = df.query(q)[column].array / 1000

        mid = (len(labels) / 2)
        offset = i - mid
        rects = ax.bar(x + (offset * .1) + .05, group_data, .1, label=l)
        groups.append(group_data)
        i += 1

    ax.set_ylabel(ylabel)
    ax.set_xlabel(xlabel)
    ax.set_xticks(x)
    ax.set_xticklabels(columns)
    ax.legend()

    plt.title('%s latency' % column)
    fig.tight_layout()

    plt.savefig('_'.join([name, column]) + ".svg")


cpu = {
    "name": "cpu",
    "xlabel": "# of conns (1000 qps total)",
    "ylabel": "mCPU",
    "columns": [
        "cpu_mili_avg_istio_proxy_fortioclient",
        "cpu_mili_avg_istio_proxy_fortioserver",
    ]
}

mem = {
    "name": "mem",
    "xlabel": "# of conns (1000 qps total)",
    "ylabel": "latency (ms)",
    "columns": [
        "cpu_mili_avg_istio_proxy_fortioclient",
        "cpu_mili_avg_istio_proxy_fortioserver",
    ]
}
latencies = {
    "name": "latencies",
    "xlabel": "# of conns (1000 qps total)",
    "ylabel": "latency (ms)",
    "columns": [
        "p50",
        "p90",
        "p99",
        "p999",
    ]
}

mode = latencies

for column in mode["columns"]:
    build_chart(column, mode["xlabel"], mode["ylabel"], mode["name"])

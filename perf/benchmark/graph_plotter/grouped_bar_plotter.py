import matplotlib.pyplot as plt
import numpy as np
import pandas as pd

df = pd.read_csv('grpcbench.csv')


def build_chart(percentile):
    labels = ['2', '4', '8', '16', '32', '64']
    suffixes = [
        # 'baseline',
        # 'grpcxds_to_proxyless',
        'grpcxds_to_envoy_plaintext',
        'grpcxds_to_envoy_mtls',
        'envoy_envoy_mtls_disable',
        'envoy_envoy_mtls',
    ]

    x = np.arange(len(labels))
    fig, ax = plt.subplots()
    groups = []
    i = 0
    for s in suffixes:
        q = 'Labels=="%s"' % s
        group_data = df.query(q)[percentile].array / 1000

        mid = (len(suffixes) / 2)
        offset = i - mid
        rects = ax.bar(x + (offset * .1) + .05, group_data, .1, label=s)

        groups.append(group_data)
        i += 1

    ax.set_ylabel('latency (ms)')
    ax.set_xlabel('# of conns (1000 qps each)')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend()

    plt.title('%s latency' % percentile)
    fig.tight_layout()

    plt.show()


percentiles = [
    "p50",
    # "p90",
    "p99",
    # "p999",
]

for percentile in percentiles:
    build_chart(percentile)

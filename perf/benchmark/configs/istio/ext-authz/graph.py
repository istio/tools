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
import sys


if __name__ == '__main__':
    df = pd.read_csv(sys.argv[1])

    # parse data
    data = [{}, {}, {}, {}]
    for _, row in df.iterrows():
        label = row['Labels']
        mode = label.split('_')[4]
        for i in range(len(data)):
            if mode not in data[i]:
                data[i][mode] = [0, 0, 0]

        idx = 0
        if label.split('_')[5] == 'medium':
            idx = 1
        if label.split('_')[5] == 'large':
            idx = 2

        for i, p in zip(range(4), ['p50', 'p90', 'p99', 'p999']):
            data[i][mode][idx] = float(row[p]) / 1000.0

    # draw graphs
    x = [8, 32, 64]
    dpi = 100
    for i, p in zip(range(4), ['p50', 'p90', 'p99', 'p999']):
        plt.figure(figsize=(1138 / dpi, 871 / dpi), dpi=dpi)
        plt.plot(x, data[i]['with-ext-authz'], 'b', label='to workload with ext-authz ' + p, marker='o')
        plt.plot(x, data[i]['without-ext-authz'], 'g', label='to workload without ext-authz ' + p, marker='o')
        plt.plot(x, data[i]['to-ext-authz'], 'y', label='to ext-authz provider ' + p, marker='o')
        plt.legend()
        plt.grid()
        plt.xlabel('connections')
        plt.ylabel('latency, milliseconds')
        plt.savefig('results/' + p + '.png', dpi=dpi)

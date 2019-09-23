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

import sys
import argparse
import itertools  # for cycling through colors
import pandas as pd
from bokeh.plotting import figure, output_file, show, save
from bokeh.palettes import Dark2_5 as palette


# generate_chart displays numthreads vs. metric, writes to interactive HTML
def generate_chart(mesh, csv, x_label, y_label_short, charts_output, show_graph):
    print(
        "Generating chart, x_label=%s, y_label=%s, csv=%s ..." %
        (x_label, y_label_short, csv))

    # valid options = latency, memory, cpu
    valid_metrics = {"p50": "p50",
                     "p90": "p90",
                     "p99": "p99",
                     "mem": "mem_MB_max_fortioserver_deployment_proxy",
                     "cpu": "cpu_mili_max_fortioserver_deployment_proxy"}

    if y_label_short is None:
        sys.exit('need metric')
    if y_label_short not in valid_metrics:
        sys.exit("invalid metric")
    if csv is None:
        sys.exit('need CSV file')

    y_label = valid_metrics[y_label_short]  # the CSV label

    # 1. read CSV to pandas dataframe
    df = pd.read_csv(csv, index_col=None, header=0)
    df["Labels"] = [x.split('_', 6)[-1] for x in df['Labels']]

    # 2. generate series to plot (x= numthreads or qps, y=y_)
    x_series, y_series = get_series(df, x_label, y_label)

    # 3. generate title
    qps = df.at[0, 'ActualQPS']
    seconds = df.at[0, 'ActualDuration']
    threads = df.at[0, 'NumThreads']

    if x_label == "connections":  # option 1 -- constant QPS, numthreads x metric
        title = "{} {}, {} QPS over {} seconds".format(
            mesh, y_label_short, qps, seconds)
    else:  # option 2 -- constant numthreads, QPS x metric
        title = "{} {}, {} threads over {} seconds".format(
            mesh, y_label_short, threads, seconds)

    # 4. prep file-write
    if charts_output == "":
        fn = "".join(title.split())
        charts_output = "/tmp/" + fn + ".html"
    output_file(charts_output)

    # 5. create chart -> save as interactive HTML
    p = build_chart(title, x_label, x_series, y_label, y_label_short, y_series)
    if show_graph:
        show(p)
    else:
        save(p)
        print("HTML graph saved at %s" % charts_output)


# get_series processes x_label metric / y-axis metric for different test
# modes (both, serveronly, etc.)
def get_series(df, x_label, metric):

    # map CSV label regex --> cleaner labels for graph legend
    modes = [('^serveronly', 'serveronly'),
             ("nomix.*_serveronly", "nomixer_serveronly"),
             ("nomix.*_both", "nomixer_both"),
             ("base", "base"),
             # Make sure this set of data is last, because both type always have data to make sure rows are not empty.
             ("^both", "both")]

    # get y axis
    series = {}
    rows = pd.DataFrame()
    for mode in modes:
        temp_row = df[df.Labels.str.contains(mode[0])]
        if temp_row.size == 0:
            continue
        rows = temp_row
        vals = list(rows[metric])

        # if y-axis metric is latency, convert microseconds to milliseconds
        if metric.startswith("p"):
            print("converting CSV microseconds to milliseconds...")
            vals = [v / 1000 for v in vals]
        # reverse to match sorted numthreads, below
        vals.reverse()
        series[mode[1]] = vals

    # only include test modes that were in the input CSV - (if nomixer not
    # included, don't attempt to plot it)
    useries = {}
    for k, v in series.items():
        if v:
            useries[k] = v
    y = useries

    # get x axis
    if rows.size == 0:
        print("warn: size of x axis dataframe is 0")
    if x_label == "connections":
        x = list(rows.NumThreads)
    elif x_label == "qps":
        x = list(rows.ActualQPS)
    else:
        sys.exit("Error: x_label must be one of: connections,qps")

    x.sort()  # sort to increasing order
    return x, y


# build_chart creates a bokeh.js plot from data
def build_chart(title, x_label, x_series, y_label, y_label_short, y_series):
    # generate y-axis label with units
    print(y_label)
    if y_label.startswith('p'):
        y_axis_label = " latency, milliseconds"
    else:
        if y_label.startswith('mem'):
            y_axis_label = "max memory usage, server proxy (MB)"
        else:
            y_axis_label = "max CPUs, server proxy (millicores)"

    # show metric value on hover
    TOOLTIPS = [(y_label_short, '$data_y')]

    # create plot
    p = figure(
        tools="pan,box_zoom,reset,save",
        title=title,
        tooltips=TOOLTIPS,
        plot_width=1000, plot_height=600,
        x_axis_label=x_label, y_axis_label=y_axis_label
    )

    # format axes
    p.title.text_font_size = '22pt'
    p.xaxis.minor_tick_line_color = None  # turn off x-axis minor ticks
    p.yaxis.minor_tick_line_color = None  # turn off y-axis minor ticks
    p.xaxis.axis_label_text_font_size = "15pt"
    p.yaxis.axis_label_text_font_size = "15pt"
    p.xaxis.major_label_text_font_size = "13pt"
    p.yaxis.major_label_text_font_size = "13pt"
    p.xaxis.ticker = x_series  # x (qps, numthreads) is a discrete variable

    # use a different color for each series (both, baseline, etc.)
    colors = itertools.cycle(palette)
    for mode, val in y_series.items():
        col = next(colors)
        p.line(x_series, val, line_color=col)
        p.circle(x_series, val, legend=mode, size=10, fill_color=col)

    p.legend.location = "top_left"

    return p


def main(argv):
    args = get_parser().parse_args(argv)
    return generate_chart(args.mesh, args.csv, args.xaxis,
                          args.metric, args.charts_output, args.show_graph)


def get_parser():
    parser = argparse.ArgumentParser(
        "Service Mesh Performance Graph Generator")
    parser.add_argument("csv", help="csv file", default="")
    parser.add_argument(
        "metric",
        help="y-axis: one of: p50, p90, p99, mem, cpu",
        default="")
    parser.add_argument(
        "--xaxis",
        help="one of: connections, qps",
        default="connections")
    parser.add_argument(
        "--charts_output",
        help="output path of generated charts",
        default=""
    )
    parser.add_argument(
        "--mesh",
        help="which service mesh tool: istio, linkerd",
        default="istio")

    parser.add_argument(
        "--show_graph",
        help="whether to show graph in browser or just save it",
        action="store_true",
    )
    return parser


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

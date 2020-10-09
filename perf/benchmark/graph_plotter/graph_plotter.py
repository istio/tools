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
import pandas as pd
import matplotlib.pyplot as plt


metric_dict = {"cpu-client": "cpu_mili_avg_istio_proxy_fortioclient",
               "cpu-server": "cpu_mili_avg_istio_proxy_fortioserver",
               "mem-client": "mem_Mi_avg_istio_proxy_fortioclient",
               "mem-server": "mem_Mi_avg_istio_proxy_fortioserver"}


def plotter(args):
    check_if_args_provided(args)

    df = pd.read_csv(args.csv_filepath)
    telemetry_modes_y_data = {}
    metric_name = get_metric_name(args)
    constructed_query_str = get_constructed_query_str(args)

    for telemetry_mode in args.telemetry_modes:
        telemetry_modes_y_data[telemetry_mode] = get_data_helper(df, args.query_list, constructed_query_str,
                                                                 telemetry_mode, metric_name)

    dpi = 100
    plt.figure(figsize=(1138 / dpi, 871 / dpi), dpi=dpi)
    for key, val in telemetry_modes_y_data.items():
        plt.plot(args.query_list, val, marker='o', label=key)

    plt.xlabel(get_x_label(args))
    plt.ylabel(get_y_label(args))
    plt.legend()
    plt.grid()
    plt.savefig(args.graph_title, dpi=dpi)
    plt.show()


# Helpers
def check_if_args_provided(args):
    args_all_provided = True
    # print(vars(args))
    for _, val in vars(args).items():
        if val == "":
            print("Warning: There is at least one argument that you did not specify with a value.\n")
            args_all_provided = False
    if not check_args_consistency(args):
        args_all_provided = False
    if not args_all_provided:
        sys.exit(-1)


def check_args_consistency(args):
    if args.x_axis == "conn" and not args.query_str.startswith("ActualQPS=="):
        print("Warning: your specified query_str does not match with the x_axis definition.")
        return False
    if args.x_axis == "qps" and not args.query_str.startswith("NumThreads=="):
        print("Warning: your specified query_str does not match with the x_axis definition.")
        return False
    return True


def get_constructed_query_str(args):
    if args.x_axis == "qps":
        return 'ActualQPS==@ql and ' + args.query_str + ' and Labels.str.endswith(@telemetry_mode)'
    elif args.x_axis == "conn":
        return args.query_str + ' and NumThreads==@ql and Labels.str.endswith(@telemetry_mode)'
    return ""


def get_metric_name(args):
    if args.graph_type.startswith("latency"):
        return args.graph_type.split("-")[1]
    return metric_dict[args.graph_type]


def get_data_helper(df, query_list, query_str, telemetry_mode, metric_name):
    y_series_data = []

    for ql in query_list:
        data = df.query(query_str)
        try:
            data[metric_name].head().empty
        except KeyError as e:
            y_series_data.append(None)
        else:
            if not data[metric_name].head().empty:
                if metric_name in ['cpu', 'mem']:
                    y_series_data.append(data[metric_name].head(1).values[0])
                else:
                    y_series_data.append(data[metric_name].head(1).values[0] / 1000)
            else:
                y_series_data.append(None)

    return y_series_data


def get_x_label(args):
    if args.x_axis == "qps":
        return "QPS"
    if args.x_axis == "conn":
        return "Connections"
    return ""


def get_y_label(args):
    if args.graph_type.startswith("latency"):
        return 'Latency in milliseconds'
    if args.graph_type.startswith("cpu"):
        return 'istio-proxy average CPUs (milliseconds)'
    if args.graph_type.startswith("mem"):
        return "istio-proxy average Memory (Mi)"
    return ""


def int_list(lst):
    return [int(i) for i in lst.split(",")]


def string_list(lst):
    return [str(i) for i in lst.split(",")]


def get_parser():
    parser = argparse.ArgumentParser(
        "Istio performance benchmark CSV file graph plotter.")
    parser.add_argument(
        "--graph_type",
        help="Choose from one of them: [latency-p50, latency-p90, latency-p99, latency-p999, "
             "cpu-client, cpu-server, mem-client, mem-server]."
    )
    parser.add_argument(
        "--x_axis",
        help="Either qps or conn.",
    )
    parser.add_argument(
        "--telemetry_modes",
        help="This is a list of perf test labels, currently it can be any combinations from the follow supported modes:"
             "[none_mtls_baseline, none_mtls_both, v2-sd-full-nullvm_both, v2-stats-nullvm_both, "
             "v2-stats-wasm_both, v2-sd-nologging-nullvm_both].",
        type=string_list
    )
    parser.add_argument(
        "--query_list",
        help="Specify the qps or conn range you want to plot based on the CSV file."
             "For example, conn_query_list=[2, 4, 8, 16, 32, 64], qps_query_list=[10, 100, 200, 400, 800, 1000].",
        type=int_list
    )
    parser.add_argument(
        "--query_str",
        help="Specify the qps or conn query_str that will be used to query your y-axis data based on the CSV file."
             "For example: conn_query_str=ActualQPS==1000, qps_query_str=NumThreads==16."
    )
    parser.add_argument(
        "--csv_filepath",
        help="The path of the CSV file."
    )
    parser.add_argument(
        "--graph_title",
        help="The graph title."
    )
    return parser


def main(argv):
    args = get_parser().parse_args(argv)
    return plotter(args)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

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
from functools import reduce

from django.shortcuts import render
import pandas as pd
from helpers import download
import os

cwd = os.getcwd()
perf_data_path = cwd + "/perf_data/"
current_release = [os.getenv('CUR_RELEASE')]

cur_selected_release = []
master_selected_release = []

cpu_cur_selected_release = []
cpu_master_selected_release = []
cpu_client_metric_name = 'cpu_mili_avg_istio_proxy_fortioclient'
cpu_server_metric_name = 'cpu_mili_avg_istio_proxy_fortioserver'

mem_cur_selected_release = []
mem_master_selected_release = []
mem_client_metric_name = 'mem_Mi_avg_istio_proxy_fortioclient'
mem_server_metric_name = 'mem_Mi_avg_istio_proxy_fortioserver'


def benchmarks_overview(request):
    return render(request, "benchmarks_overview.html")


# Create your views here.
def latency_vs_conn(request, uploaded_csv_url=None):
    if uploaded_csv_url is not None:
        uploaded_csv_path = cwd + uploaded_csv_url
        df = pd.read_csv(uploaded_csv_path)
        context = get_lantency_vs_conn_context(df)
        os.remove(uploaded_csv_path)
        return context
    else:
        cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, \
            master_release_dates = download.download_benchmark_csv(60)
        cur_benchmark_test_ids = get_benchmark_test_ids(cur_href_links)
        master_benchmark_test_ids = get_benchmark_test_ids(master_href_links)

        if request.method == "POST" and 'current_release_name' in request.POST:
            cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(cur_selected_release) > 1:
            cur_selected_release.pop(0)
        if len(cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")

        release_context = get_lantency_vs_conn_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(master_selected_release) > 1:
            master_selected_release.pop(0)
        if len(master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")

        latency_none_mtls_base_p50_master = get_latency_vs_conn_y_series(df, '_none_mtls_baseline', 'p50')
        latency_none_mtls_both_p50_master = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p50')
        latency_none_plaintext_both_p50_master = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p50')
        latency_v2_stats_nullvm_both_p50_master = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p50')
        latency_v2_stats_wasm_both_p50_master = get_latency_vs_conn_y_series(df, '_v2-stats-wasm_both', 'p50')
        latency_v2_sd_nologging_nullvm_both_p50_master = get_latency_vs_conn_y_series(df,
                                                                                      '_v2-sd-nologging-nullvm_both',
                                                                                      'p50')
        latency_v2_sd_full_nullvm_both_p50_master = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p50')

        latency_none_mtls_base_p90_master = get_latency_vs_conn_y_series(df, '_none_mtls_baseline', 'p90')
        latency_none_mtls_both_p90_master = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p90')
        latency_none_plaintext_both_p90_master = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p90')
        latency_v2_stats_nullvm_both_p90_master = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p90')
        latency_v2_stats_wasm_both_p90_master = get_latency_vs_conn_y_series(df, '_v2-stats-wasm_both', 'p90')
        latency_v2_sd_nologging_nullvm_both_p90_master = get_latency_vs_conn_y_series(df,
                                                                                      '_v2-sd-nologging-nullvm_both',
                                                                                      'p90')
        latency_v2_sd_full_nullvm_both_p90_master = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p90')

        latency_none_mtls_base_p99_master = get_latency_vs_conn_y_series(df, '_none_mtls_baseline', 'p99')
        latency_none_mtls_both_p99_master = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p99')
        latency_none_plaintext_both_p99_master = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p99')
        latency_v2_stats_nullvm_both_p99_master = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p99')
        latency_v2_stats_wasm_both_p99_master = get_latency_vs_conn_y_series(df, '_v2-stats-wasm_both', 'p99')
        latency_v2_sd_nologging_nullvm_both_p99_master = get_latency_vs_conn_y_series(df,
                                                                                      '_v2-sd-nologging-nullvm_both',
                                                                                      'p99')
        latency_v2_sd_full_nullvm_both_p99_master = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p99')

        latency_none_mtls_base_p999_master = get_latency_vs_conn_y_series(df, '_none_mtls_baseline', 'p99.9')
        latency_none_mtls_both_p999_master = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p99.9')
        latency_none_plaintext_both_p999_master = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p99.9')
        latency_v2_stats_nullvm_both_p999_master = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p99.9')
        latency_v2_stats_wasm_both_p999_master = get_latency_vs_conn_y_series(df, '_v2-stats-wasm_both', 'p99.9')
        latency_v2_sd_nologging_nullvm_both_p999_master = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99.9')
        latency_v2_sd_full_nullvm_both_p999_master = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p99.9')

        other_context = {'current_release': current_release,
                         'cur_selected_release': cur_selected_release,
                         'master_selected_release': master_selected_release,
                         'cur_release_names': cur_benchmark_test_ids,
                         'master_release_names': master_benchmark_test_ids,
                         }

        master_context = {'latency_none_mtls_base_p50_master': latency_none_mtls_base_p50_master,
                          'latency_none_mtls_both_p50_master': latency_none_mtls_both_p50_master,
                          'latency_none_plaintext_both_p50_master': latency_none_plaintext_both_p50_master,
                          'latency_v2_stats_nullvm_both_p50_master': latency_v2_stats_nullvm_both_p50_master,
                          'latency_v2_stats_wasm_both_p50_master': latency_v2_stats_wasm_both_p50_master,
                          'latency_v2_sd_nologging_nullvm_both_p50_master': latency_v2_sd_nologging_nullvm_both_p50_master,
                          'latency_v2_sd_full_nullvm_both_p50_master': latency_v2_sd_full_nullvm_both_p50_master,
                          'latency_none_mtls_base_p90_master': latency_none_mtls_base_p90_master,
                          'latency_none_mtls_both_p90_master': latency_none_mtls_both_p90_master,
                          'latency_none_plaintext_both_p90_master': latency_none_plaintext_both_p90_master,
                          'latency_v2_stats_nullvm_both_p90_master': latency_v2_stats_nullvm_both_p90_master,
                          'latency_v2_stats_wasm_both_p90_master': latency_v2_stats_wasm_both_p90_master,
                          'latency_v2_sd_nologging_nullvm_both_p90_master': latency_v2_sd_nologging_nullvm_both_p90_master,
                          'latency_v2_sd_full_nullvm_both_p90_master': latency_v2_sd_full_nullvm_both_p90_master,
                          'latency_none_mtls_base_p99_master': latency_none_mtls_base_p99_master,
                          'latency_none_mtls_both_p99_master': latency_none_mtls_both_p99_master,
                          'latency_none_plaintext_both_p99_master': latency_none_plaintext_both_p99_master,
                          'latency_v2_stats_nullvm_both_p99_master': latency_v2_stats_nullvm_both_p99_master,
                          'latency_v2_stats_wasm_both_p99_master': latency_v2_stats_wasm_both_p99_master,
                          'latency_v2_sd_nologging_nullvm_both_p99_master': latency_v2_sd_nologging_nullvm_both_p99_master,
                          'latency_v2_sd_full_nullvm_both_p99_master': latency_v2_sd_full_nullvm_both_p99_master,
                          'latency_none_mtls_base_p999_master': latency_none_mtls_base_p999_master,
                          'latency_none_mtls_both_p999_master': latency_none_mtls_both_p999_master,
                          'latency_none_plaintext_both_p999_master': latency_none_plaintext_both_p999_master,
                          'latency_v2_stats_nullvm_both_p999_master': latency_v2_stats_nullvm_both_p999_master,
                          'latency_v2_stats_wasm_both_p999_master': latency_v2_stats_wasm_both_p999_master,
                          'latency_v2_sd_nologging_nullvm_both_p999_master': latency_v2_sd_nologging_nullvm_both_p999_master,
                          'latency_v2_sd_full_nullvm_both_p999_master': latency_v2_sd_full_nullvm_both_p999_master,
                          }

        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))

        return render(request, "latency_vs_conn.html", context=context)


def get_benchmark_test_ids(href_links):
    benchmark_test_ids = []
    for link in href_links:
        benchmark_test_ids.append(link.split("/")[4])
    return benchmark_test_ids


def latency_vs_qps(request, uploaded_csv_url=None):
    if uploaded_csv_url is not None:
        uploaded_csv_path = cwd + uploaded_csv_url
        df = pd.read_csv(uploaded_csv_path)
        context = get_lantency_vs_qps_context(df)
        os.remove(uploaded_csv_path)
        return context
    else:
        cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, \
            master_release_dates = download.download_benchmark_csv(60)
        cur_benchmark_test_ids = get_benchmark_test_ids(cur_href_links)
        master_benchmark_test_ids = get_benchmark_test_ids(master_href_links)

        if request.method == "POST" and 'current_release_name' in request.POST:
            cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(cur_selected_release) > 1:
            cur_selected_release.pop(0)
        if len(cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")

        release_context = get_lantency_vs_qps_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(master_selected_release) > 1:
            master_selected_release.pop(0)
        if len(master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")

        latency_none_mtls_base_p50_master = get_latency_vs_qps_y_series(df, '_none_mtls_baseline', 'p50')
        latency_none_mtls_both_p50_master = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p50')
        latency_none_plaintext_both_p50_master = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p50')
        latency_v2_stats_nullvm_both_p50_master = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p50')
        latency_v2_stats_wasm_both_p50_master = get_latency_vs_qps_y_series(df, '_v2-stats-wasm_both', 'p50')
        latency_v2_sd_nologging_nullvm_both_p50_master = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both',
                                                                                     'p50')
        latency_v2_sd_full_nullvm_both_p50_master = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p50')

        latency_none_mtls_base_p90_master = get_latency_vs_qps_y_series(df, '_none_mtls_baseline', 'p90')
        latency_none_mtls_both_p90_master = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p90')
        latency_none_plaintext_both_p90_master = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p90')
        latency_v2_stats_nullvm_both_p90_master = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p90')
        latency_v2_stats_wasm_both_p90_master = get_latency_vs_qps_y_series(df, '_v2-stats-wasm_both', 'p90')
        latency_v2_sd_nologging_nullvm_both_p90_master = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p90')
        latency_v2_sd_full_nullvm_both_p90_master = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p90')

        latency_none_mtls_base_p99_master = get_latency_vs_qps_y_series(df, '_none_mtls_baseline', 'p99')
        latency_none_mtls_both_p99_master = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p99')
        latency_none_plaintext_both_p99_master = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p99')
        latency_v2_stats_nullvm_both_p99_master = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p99')
        latency_v2_stats_wasm_both_p99_master = get_latency_vs_qps_y_series(df, '_v2-stats-wasm_both', 'p99')
        latency_v2_sd_nologging_nullvm_both_p99_master = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99')
        latency_v2_sd_full_nullvm_both_p99_master = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p99')

        latency_none_mtls_base_p999_master = get_latency_vs_qps_y_series(df, '_none_mtls_baseline', 'p99.9')
        latency_none_mtls_both_p999_master = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p99.9')
        latency_none_plaintext_both_p999_master = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p99.9')
        latency_v2_stats_nullvm_both_p999_master = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p99.9')
        latency_v2_stats_wasm_both_p999_master = get_latency_vs_qps_y_series(df, '_v2-stats-wasm_both', 'p99.9')
        latency_v2_sd_nologging_nullvm_both_p999_master = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99.9')
        latency_v2_sd_full_nullvm_both_p999_master = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p99.9')

        other_context = {'current_release': current_release,
                         'cur_selected_release': cur_selected_release,
                         'master_selected_release': master_selected_release,
                         'cur_release_names': cur_benchmark_test_ids,
                         'master_release_names': master_benchmark_test_ids,
                         }

        master_context = {'latency_none_mtls_base_p50_master': latency_none_mtls_base_p50_master,
                          'latency_none_mtls_both_p50_master': latency_none_mtls_both_p50_master,
                          'latency_none_plaintext_both_p50_master': latency_none_plaintext_both_p50_master,
                          'latency_v2_stats_nullvm_both_p50_master': latency_v2_stats_nullvm_both_p50_master,
                          'latency_v2_stats_wasm_both_p50_master': latency_v2_stats_wasm_both_p50_master,
                          'latency_v2_sd_nologging_nullvm_both_p50_master': latency_v2_sd_nologging_nullvm_both_p50_master,
                          'latency_v2_sd_full_nullvm_both_p50_master': latency_v2_sd_full_nullvm_both_p50_master,
                          'latency_none_mtls_base_p90_master': latency_none_mtls_base_p90_master,
                          'latency_none_mtls_both_p90_master': latency_none_mtls_both_p90_master,
                          'latency_none_plaintext_both_p90_master': latency_none_plaintext_both_p90_master,
                          'latency_v2_stats_nullvm_both_p90_master': latency_v2_stats_nullvm_both_p90_master,
                          'latency_v2_stats_wasm_both_p90_master': latency_v2_stats_wasm_both_p90_master,
                          'latency_v2_sd_nologging_nullvm_both_p90_master': latency_v2_sd_nologging_nullvm_both_p90_master,
                          'latency_v2_sd_full_nullvm_both_p90_master': latency_v2_sd_full_nullvm_both_p90_master,
                          'latency_none_mtls_base_p99_master': latency_none_mtls_base_p99_master,
                          'latency_none_mtls_both_p99_master': latency_none_mtls_both_p99_master,
                          'latency_none_plaintext_both_p99_master': latency_none_plaintext_both_p99_master,
                          'latency_v2_stats_nullvm_both_p99_master': latency_v2_stats_nullvm_both_p99_master,
                          'latency_v2_stats_wasm_both_p99_master': latency_v2_stats_wasm_both_p99_master,
                          'latency_v2_sd_nologging_nullvm_both_p99_master': latency_v2_sd_nologging_nullvm_both_p99_master,
                          'latency_v2_sd_full_nullvm_both_p99_master': latency_v2_sd_full_nullvm_both_p99_master,
                          'latency_none_mtls_base_p999_master': latency_none_mtls_base_p999_master,
                          'latency_none_mtls_both_p999_master': latency_none_mtls_both_p999_master,
                          'latency_none_plaintext_both_p999_master': latency_none_plaintext_both_p999_master,
                          'latency_v2_stats_nullvm_both_p999_master': latency_v2_stats_nullvm_both_p999_master,
                          'latency_v2_stats_wasm_both_p999_master': latency_v2_stats_wasm_both_p999_master,
                          'latency_v2_sd_nologging_nullvm_both_p999_master': latency_v2_sd_nologging_nullvm_both_p999_master,
                          'latency_v2_sd_full_nullvm_both_p999_master': latency_v2_sd_full_nullvm_both_p999_master,
                          }
        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))

        return render(request, "latency_vs_qps.html", context=context)


def cpu_vs_qps(request, uploaded_csv_url=None):
    if uploaded_csv_url is not None:
        uploaded_csv_path = cwd + uploaded_csv_url
        df = pd.read_csv(uploaded_csv_path)
        context = get_cpu_vs_qps_context(df)
        os.remove(uploaded_csv_path)
        return context
    else:
        cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, \
            master_release_dates = download.download_benchmark_csv(60)
        cur_benchmark_test_ids = get_benchmark_test_ids(cur_href_links)
        master_benchmark_test_ids = get_benchmark_test_ids(master_href_links)

        if request.method == "POST" and 'current_release_name' in request.POST:
            cpu_cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(cpu_cur_selected_release) > 1:
            cpu_cur_selected_release.pop(0)
        if len(cpu_cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")

        release_context = get_cpu_vs_qps_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(cpu_master_selected_release) > 1:
            cpu_master_selected_release.pop(0)
        if len(cpu_master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")

        cpu_client_none_mtls_base_master = get_cpu_vs_qps_y_series(df, '_none_mtls_baseline', cpu_client_metric_name)
        cpu_client_none_mtls_both_master = get_cpu_vs_qps_y_series(df, '_none_mtls_both', cpu_client_metric_name)
        cpu_client_none_plaintext_both_master = get_cpu_vs_qps_y_series(df, '_none_plaintext_both', cpu_client_metric_name)
        cpu_client_v2_stats_nullvm_both_master = get_cpu_vs_qps_y_series(df, '_v2-stats-nullvm_both', cpu_client_metric_name)
        cpu_client_v2_stats_wasm_both_master = get_cpu_vs_qps_y_series(df, '_v2-stats-wasm_both', cpu_client_metric_name)
        cpu_client_v2_sd_nologging_nullvm_both_master = get_cpu_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', cpu_client_metric_name)
        cpu_client_v2_sd_full_nullvm_both_master = get_cpu_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', cpu_client_metric_name)

        cpu_server_none_mtls_base_master = get_cpu_vs_qps_y_series(df, '_none_mtls_baseline', cpu_server_metric_name)
        cpu_server_none_mtls_both_master = get_cpu_vs_qps_y_series(df, '_none_mtls_both', cpu_server_metric_name)
        cpu_server_none_plaintext_both_master = get_cpu_vs_qps_y_series(df, '_none_plaintext_both', cpu_server_metric_name)
        cpu_server_v2_stats_nullvm_both_master = get_cpu_vs_qps_y_series(df, '_v2-stats-nullvm_both', cpu_server_metric_name)
        cpu_server_v2_stats_wasm_both_master = get_cpu_vs_qps_y_series(df, '_v2-stats-wasm_both', cpu_server_metric_name)
        cpu_server_v2_sd_nologging_nullvm_both_master = get_cpu_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', cpu_server_metric_name)
        cpu_server_v2_sd_full_nullvm_both_master = get_cpu_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', cpu_server_metric_name)

        other_context = {'current_release': current_release,
                         'cpu_cur_selected_release': cpu_cur_selected_release,
                         'cpu_master_selected_release': cpu_master_selected_release,
                         'cur_release_names': cur_benchmark_test_ids,
                         'master_release_names': master_benchmark_test_ids,
                         }

        master_context = {'cpu_client_none_mtls_base_master': cpu_client_none_mtls_base_master,
                          'cpu_client_none_mtls_both_master': cpu_client_none_mtls_both_master,
                          'cpu_client_none_plaintext_both_master': cpu_client_none_plaintext_both_master,
                          'cpu_client_v2_stats_nullvm_both_master': cpu_client_v2_stats_nullvm_both_master,
                          'cpu_client_v2_stats_wasm_both_master': cpu_client_v2_stats_wasm_both_master,
                          'cpu_client_v2_sd_nologging_nullvm_both_master': cpu_client_v2_sd_nologging_nullvm_both_master,
                          'cpu_client_v2_sd_full_nullvm_both_master': cpu_client_v2_sd_full_nullvm_both_master,
                          'cpu_server_none_mtls_base_master': cpu_server_none_mtls_base_master,
                          'cpu_server_none_mtls_both_master': cpu_server_none_mtls_both_master,
                          'cpu_server_none_plaintext_both_master': cpu_server_none_plaintext_both_master,
                          'cpu_server_v2_stats_nullvm_both_master': cpu_server_v2_stats_nullvm_both_master,
                          'cpu_server_v2_stats_wasm_both_master': cpu_server_v2_stats_wasm_both_master,
                          'cpu_server_v2_sd_nologging_nullvm_both_master': cpu_server_v2_sd_nologging_nullvm_both_master,
                          'cpu_server_v2_sd_full_nullvm_both_master': cpu_server_v2_sd_full_nullvm_both_master,
                          }

        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))
        return render(request, "cpu_vs_qps.html", context=context)


def cpu_vs_conn(request, uploaded_csv_url=None):
    if uploaded_csv_url is not None:
        uploaded_csv_path = cwd + uploaded_csv_url
        df = pd.read_csv(uploaded_csv_path)
        context = get_cpu_vs_conn_context(df)
        os.remove(uploaded_csv_path)
        return context
    else:
        cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, \
            master_release_dates = download.download_benchmark_csv(60)
        cur_benchmark_test_ids = get_benchmark_test_ids(cur_href_links)
        master_benchmark_test_ids = get_benchmark_test_ids(master_href_links)

        if request.method == "POST" and 'current_release_name' in request.POST:
            cpu_cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(cpu_cur_selected_release) > 1:
            cpu_cur_selected_release.pop(0)
        if len(cpu_cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")

        release_context = get_cpu_vs_conn_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(cpu_master_selected_release) > 1:
            cpu_master_selected_release.pop(0)
        if len(cpu_master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")

        cpu_client_none_mtls_base_master = get_cpu_vs_conn_y_series(df, '_none_mtls_baseline', cpu_client_metric_name)
        cpu_client_none_mtls_both_master = get_cpu_vs_conn_y_series(df, '_none_mtls_both', cpu_client_metric_name)
        cpu_client_none_plaintext_both_master = get_cpu_vs_conn_y_series(df, '_none_plaintext_both', cpu_client_metric_name)
        cpu_client_v2_stats_nullvm_both_master = get_cpu_vs_conn_y_series(df, '_v2-stats-nullvm_both', cpu_client_metric_name)
        cpu_client_v2_stats_wasm_both_master = get_cpu_vs_conn_y_series(df, '_v2-wasm-nullvm_both', cpu_client_metric_name)
        cpu_client_v2_sd_nologging_nullvm_both_master = get_cpu_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', cpu_client_metric_name)
        cpu_client_v2_sd_full_nullvm_both_master = get_cpu_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', cpu_client_metric_name)

        cpu_server_none_mtls_base_master = get_cpu_vs_conn_y_series(df, '_none_mtls_baseline', cpu_server_metric_name)
        cpu_server_none_mtls_both_master = get_cpu_vs_conn_y_series(df, '_none_mtls_both', cpu_server_metric_name)
        cpu_server_none_plaintext_both_master = get_cpu_vs_conn_y_series(df, '_none_plaintext_both', cpu_server_metric_name)
        cpu_server_v2_stats_nullvm_both_master = get_cpu_vs_conn_y_series(df, '_v2-stats-nullvm_both', cpu_server_metric_name)
        cpu_server_v2_stats_wasm_both_master = get_cpu_vs_conn_y_series(df, '_v2-stats-wasm_both', cpu_server_metric_name)
        cpu_server_v2_sd_nologging_nullvm_both_master = get_cpu_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', cpu_server_metric_name)
        cpu_server_v2_sd_full_nullvm_both_master = get_cpu_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', cpu_server_metric_name)

        other_context = {'current_release': current_release,
                         'cpu_cur_selected_release': cpu_cur_selected_release,
                         'cpu_master_selected_release': cpu_master_selected_release,
                         'cur_release_names': cur_benchmark_test_ids,
                         'master_release_names': master_benchmark_test_ids,
                         }

        master_context = {'cpu_client_none_mtls_base_master': cpu_client_none_mtls_base_master,
                          'cpu_client_none_mtls_both_master': cpu_client_none_mtls_both_master,
                          'cpu_client_none_plaintext_both_master': cpu_client_none_plaintext_both_master,
                          'cpu_client_v2_stats_nullvm_both_master': cpu_client_v2_stats_nullvm_both_master,
                          'cpu_client_v2_stats_wasm_both_master': cpu_client_v2_stats_wasm_both_master,
                          'cpu_client_v2_sd_nologging_nullvm_both_master': cpu_client_v2_sd_nologging_nullvm_both_master,
                          'cpu_client_v2_sd_full_nullvm_both_master': cpu_client_v2_sd_full_nullvm_both_master,
                          'cpu_server_none_mtls_base_master': cpu_server_none_mtls_base_master,
                          'cpu_server_none_mtls_both_master': cpu_server_none_mtls_both_master,
                          'cpu_server_none_plaintext_both_master': cpu_server_none_plaintext_both_master,
                          'cpu_server_v2_stats_nullvm_both_master': cpu_server_v2_stats_nullvm_both_master,
                          'cpu_server_v2_stats_wasm_both_master': cpu_server_v2_stats_wasm_both_master,
                          'cpu_server_v2_sd_nologging_nullvm_both_master': cpu_server_v2_sd_nologging_nullvm_both_master,
                          'cpu_server_v2_sd_full_nullvm_both_master': cpu_server_v2_sd_full_nullvm_both_master,
                          }

        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))
        return render(request, "cpu_vs_conn.html", context=context)


def mem_vs_qps(request, uploaded_csv_url=None):
    if uploaded_csv_url is not None:
        uploaded_csv_path = cwd + uploaded_csv_url
        df = pd.read_csv(uploaded_csv_path)
        context = get_mem_vs_qps_context(df)
        os.remove(uploaded_csv_path)
        return context
    else:
        cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, \
            master_release_dates = download.download_benchmark_csv(60)
        cur_benchmark_test_ids = get_benchmark_test_ids(cur_href_links)
        master_benchmark_test_ids = get_benchmark_test_ids(master_href_links)

        if request.method == "POST" and 'current_release_name' in request.POST:
            mem_cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(mem_cur_selected_release) > 1:
            mem_cur_selected_release.pop(0)
        if len(mem_cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")

        release_context = get_mem_vs_qps_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(mem_master_selected_release) > 1:
            mem_master_selected_release.pop(0)
        if len(mem_master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")

        mem_client_none_mtls_base_master = get_mem_vs_qps_y_series(df, '_none_mtls_baseline', mem_client_metric_name)
        mem_client_none_mtls_both_master = get_mem_vs_qps_y_series(df, '_none_mtls_both', mem_client_metric_name)
        mem_client_none_plaintext_both_master = get_mem_vs_qps_y_series(df, '_none_plaintext_both', mem_client_metric_name)
        mem_client_v2_stats_nullvm_both_master = get_mem_vs_qps_y_series(df, '_v2-stats-nullvm_both', mem_client_metric_name)
        mem_client_v2_stats_wasm_both_master = get_mem_vs_qps_y_series(df, '_v2-stats-wasm_both', mem_client_metric_name)
        mem_client_v2_sd_nologging_nullvm_both_master = get_mem_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', mem_client_metric_name)
        mem_client_v2_sd_full_nullvm_both_master = get_mem_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', mem_client_metric_name)

        mem_server_none_mtls_base_master = get_mem_vs_qps_y_series(df, '_none_mtls_baseline', mem_server_metric_name)
        mem_server_none_mtls_both_master = get_mem_vs_qps_y_series(df, '_none_mtls_both', mem_server_metric_name)
        mem_server_none_plaintext_both_master = get_mem_vs_qps_y_series(df, '_none_plaintext_both', mem_server_metric_name)
        mem_server_v2_stats_nullvm_both_master = get_mem_vs_qps_y_series(df, '_v2-stats-nullvm_both', mem_server_metric_name)
        mem_server_v2_stats_wasm_both_master = get_mem_vs_qps_y_series(df, '_v2-stats-wasm_both', mem_server_metric_name)
        mem_server_v2_sd_nologging_nullvm_both_master = get_mem_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', mem_server_metric_name)
        mem_server_v2_sd_full_nullvm_both_master = get_mem_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', mem_server_metric_name)

        other_context = {'current_release': current_release,
                         'mem_cur_selected_release': mem_cur_selected_release,
                         'mem_master_selected_release': mem_master_selected_release,
                         'cur_release_names': cur_benchmark_test_ids,
                         'master_release_names': master_benchmark_test_ids,
                         }

        master_context = {'mem_client_none_mtls_base_master': mem_client_none_mtls_base_master,
                          'mem_client_none_mtls_both_master': mem_client_none_mtls_both_master,
                          'mem_client_none_plaintext_both_master': mem_client_none_plaintext_both_master,
                          'mem_client_v2_stats_nullvm_both_master': mem_client_v2_stats_nullvm_both_master,
                          'mem_client_v2_stats_wasm_both_master': mem_client_v2_stats_wasm_both_master,
                          'mem_client_v2_sd_nologging_nullvm_both_master': mem_client_v2_sd_nologging_nullvm_both_master,
                          'mem_client_v2_sd_full_nullvm_both_master': mem_client_v2_sd_full_nullvm_both_master,
                          'mem_server_none_mtls_base_master': mem_server_none_mtls_base_master,
                          'mem_server_none_mtls_both_master': mem_server_none_mtls_both_master,
                          'mem_server_none_plaintext_both_master': mem_server_none_plaintext_both_master,
                          'mem_server_v2_stats_nullvm_both_master': mem_server_v2_stats_nullvm_both_master,
                          'mem_server_v2_stats_wasm_both_master': mem_server_v2_stats_wasm_both_master,
                          'mem_server_v2_sd_nologging_nullvm_both_master': mem_server_v2_sd_nologging_nullvm_both_master,
                          'mem_server_v2_sd_full_nullvm_both_master': mem_server_v2_sd_full_nullvm_both_master,
                          }

        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))
        return render(request, "mem_vs_qps.html", context=context)


def mem_vs_conn(request, uploaded_csv_url=None):
    if uploaded_csv_url is not None:
        uploaded_csv_path = cwd + uploaded_csv_url
        df = pd.read_csv(uploaded_csv_path)
        context = get_mem_vs_conn_context(df)
        os.remove(uploaded_csv_path)
        return context
    else:
        cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, \
            master_release_dates = download.download_benchmark_csv(60)
        cur_benchmark_test_ids = get_benchmark_test_ids(cur_href_links)
        master_benchmark_test_ids = get_benchmark_test_ids(master_href_links)

        if request.method == "POST" and 'current_release_name' in request.POST:
            mem_cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(mem_cur_selected_release) > 1:
            mem_cur_selected_release.pop(0)
        if len(mem_cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cur_href_links[0].split("/")[4] + "_benchmark.csv")

        release_context = get_mem_vs_conn_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")
        # Parse data for the current release
        if len(mem_master_selected_release) > 1:
            mem_master_selected_release.pop(0)
        if len(mem_master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + master_href_links[0].split("/")[4] + "_benchmark.csv")

        mem_client_none_mtls_base_master = get_mem_vs_conn_y_series(df, '_none_mtls_baseline', mem_client_metric_name)
        mem_client_none_mtls_both_master = get_mem_vs_conn_y_series(df, '_none_mtls_both', mem_client_metric_name)
        mem_client_none_plaintext_both_master = get_mem_vs_conn_y_series(df, '_none_plaintext_both', mem_client_metric_name)
        mem_client_v2_stats_nullvm_both_master = get_mem_vs_conn_y_series(df, '_v2-stats-nullvm_both', mem_client_metric_name)
        mem_client_v2_stats_wasm_both_master = get_mem_vs_conn_y_series(df, '_v2-stats-wasm_both', mem_client_metric_name)
        mem_client_v2_sd_nologging_nullvm_both_master = get_mem_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', mem_client_metric_name)
        mem_client_v2_sd_full_nullvm_both_master = get_mem_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', mem_client_metric_name)

        mem_server_none_mtls_base_master = get_mem_vs_conn_y_series(df, '_none_mtls_baseline', mem_server_metric_name)
        mem_server_none_mtls_both_master = get_mem_vs_conn_y_series(df, '_none_mtls_both', mem_server_metric_name)
        mem_server_none_plaintext_both_master = get_mem_vs_conn_y_series(df, '_none_plaintext_both', mem_server_metric_name)
        mem_server_v2_stats_nullvm_both_master = get_mem_vs_conn_y_series(df, '_v2-stats-nullvm_both', mem_server_metric_name)
        mem_server_v2_stats_wasm_both_master = get_mem_vs_conn_y_series(df, '_v2-stats-wasm_both', mem_server_metric_name)
        mem_server_v2_sd_nologging_nullvm_both_master = get_mem_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', mem_server_metric_name)
        mem_server_v2_sd_full_nullvm_both_master = get_mem_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', mem_server_metric_name)

        other_context = {'current_release': current_release,
                         'mem_cur_selected_release': mem_cur_selected_release,
                         'mem_master_selected_release': mem_master_selected_release,
                         'cur_release_names': cur_benchmark_test_ids,
                         'master_release_names': master_benchmark_test_ids,
                         }

        master_context = {'mem_client_none_mtls_base_master': mem_client_none_mtls_base_master,
                          'mem_client_none_mtls_both_master': mem_client_none_mtls_both_master,
                          'mem_client_none_plaintext_both_master': mem_client_none_plaintext_both_master,
                          'mem_client_v2_stats_nullvm_both_master': mem_client_v2_stats_nullvm_both_master,
                          'mem_client_v2_stats_wasm_both_master': mem_client_v2_stats_wasm_both_master,
                          'mem_client_v2_sd_nologging_nullvm_both_master': mem_client_v2_sd_nologging_nullvm_both_master,
                          'mem_client_v2_sd_full_nullvm_both_master': mem_client_v2_sd_full_nullvm_both_master,
                          'mem_server_none_mtls_base_master': mem_server_none_mtls_base_master,
                          'mem_server_none_mtls_both_master': mem_server_none_mtls_both_master,
                          'mem_server_none_plaintext_both_master': mem_server_none_plaintext_both_master,
                          'mem_server_v2_stats_nullvm_both_master': mem_server_v2_stats_nullvm_both_master,
                          'mem_server_v2_stats_wasm_both_master': mem_server_v2_stats_wasm_both_master,
                          'mem_server_v2_sd_nologging_nullvm_both_master': mem_server_v2_sd_nologging_nullvm_both_master,
                          'mem_server_v2_sd_full_nullvm_both_master': mem_server_v2_sd_full_nullvm_both_master,
                          }

        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))
        return render(request, "mem_vs_conn.html", context=context)


def get_lantency_vs_conn_context(df):
    latency_none_mtls_base_p50 = get_latency_vs_conn_y_series(df, '_none_mtls_baseline', 'p50')
    latency_none_mtls_both_p50 = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p50')
    latency_none_plaintext_both_p50 = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p50')
    latency_v2_stats_nullvm_both_p50 = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p50')
    latency_v2_stats_wasm_both_p50 = get_latency_vs_conn_y_series(df, '_v2-stats-wasm_both', 'p50')
    latency_v2_sd_nologging_nullvm_both_p50 = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p50')
    latency_v2_sd_full_nullvm_both_p50 = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p50')

    latency_none_mtls_base_p90 = get_latency_vs_conn_y_series(df, '_none_mtls_baseline', 'p90')
    latency_none_mtls_both_p90 = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p90')
    latency_none_plaintext_both_p90 = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p90')
    latency_v2_stats_nullvm_both_p90 = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p90')
    latency_v2_stats_wasm_both_p90 = get_latency_vs_conn_y_series(df, '_v2-stats-wasm_both', 'p90')
    latency_v2_sd_nologging_nullvm_both_p90 = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p90')
    latency_v2_sd_full_nullvm_both_p90 = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p90')

    latency_none_mtls_base_p99 = get_latency_vs_conn_y_series(df, '_none_mtls_baseline', 'p99')
    latency_none_mtls_both_p99 = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p99')
    latency_none_plaintext_both_p99 = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p99')
    latency_v2_stats_nullvm_both_p99 = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p99')
    latency_v2_stats_wasm_both_p99 = get_latency_vs_conn_y_series(df, '_v2-stats-wasm_both', 'p99')
    latency_v2_sd_nologging_nullvm_both_p99 = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99')
    latency_v2_sd_full_nullvm_both_p99 = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p99')

    latency_none_mtls_base_p999 = get_latency_vs_conn_y_series(df, '_none_mtls_baseline', 'p99.9')
    latency_none_mtls_both_p999 = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p99.9')
    latency_none_plaintext_both_p999 = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p99.9')
    latency_v2_stats_nullvm_both_p999 = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p99.9')
    latency_v2_stats_wasm_both_p999 = get_latency_vs_conn_y_series(df, '_v2-stats-wasm_both', 'p99.9')
    latency_v2_sd_nologging_nullvm_both_p999 = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99.9')
    latency_v2_sd_full_nullvm_both_p999 = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p99.9')

    context = {'latency_none_mtls_base_p50': latency_none_mtls_base_p50,
               'latency_none_mtls_both_p50': latency_none_mtls_both_p50,
               'latency_none_plaintext_both_p50': latency_none_plaintext_both_p50,
               'latency_v2_stats_nullvm_both_p50': latency_v2_stats_nullvm_both_p50,
               'latency_v2_stats_wasm_both_p50': latency_v2_stats_wasm_both_p50,
               'latency_v2_sd_nologging_nullvm_both_p50': latency_v2_sd_nologging_nullvm_both_p50,
               'latency_v2_sd_full_nullvm_both_p50': latency_v2_sd_full_nullvm_both_p50,
               'latency_none_mtls_base_p90': latency_none_mtls_base_p90,
               'latency_none_mtls_both_p90': latency_none_mtls_both_p90,
               'latency_none_plaintext_both_p90': latency_none_plaintext_both_p90,
               'latency_v2_stats_nullvm_both_p90': latency_v2_stats_nullvm_both_p90,
               'latency_v2_stats_wasm_both_p90': latency_v2_stats_wasm_both_p90,
               'latency_v2_sd_nologging_nullvm_both_p90': latency_v2_sd_nologging_nullvm_both_p90,
               'latency_v2_sd_full_nullvm_both_p90': latency_v2_sd_full_nullvm_both_p90,
               'latency_none_mtls_base_p99': latency_none_mtls_base_p99,
               'latency_none_mtls_both_p99': latency_none_mtls_both_p99,
               'latency_none_plaintext_both_p99': latency_none_plaintext_both_p99,
               'latency_v2_stats_nullvm_both_p99': latency_v2_stats_nullvm_both_p99,
               'latency_v2_stats_wasm_both_p99': latency_v2_stats_wasm_both_p99,
               'latency_v2_sd_nologging_nullvm_both_p99': latency_v2_sd_nologging_nullvm_both_p99,
               'latency_v2_sd_full_nullvm_both_p99': latency_v2_sd_full_nullvm_both_p99,
               'latency_none_mtls_base_p999': latency_none_mtls_base_p999,
               'latency_none_mtls_both_p999': latency_none_mtls_both_p999,
               'latency_none_plaintext_both_p999': latency_none_plaintext_both_p999,
               'latency_v2_stats_nullvm_both_p999': latency_v2_stats_nullvm_both_p999,
               'latency_v2_stats_wasm_both_p999': latency_v2_stats_wasm_both_p999,
               'latency_v2_sd_nologging_nullvm_both_p999': latency_v2_sd_nologging_nullvm_both_p999,
               'latency_v2_sd_full_nullvm_both_p999': latency_v2_sd_full_nullvm_both_p999,
               }
    return context


def get_lantency_vs_qps_context(df):
    latency_none_mtls_base_p50 = get_latency_vs_qps_y_series(df, '_none_mtls_baseline', 'p50')
    latency_none_mtls_both_p50 = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p50')
    latency_none_plaintext_both_p50 = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p50')
    latency_v2_stats_nullvm_both_p50 = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p50')
    latency_v2_stats_wasm_both_p50 = get_latency_vs_qps_y_series(df, '_v2-stats-wasm_both', 'p50')
    latency_v2_sd_nologging_nullvm_both_p50 = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p50')
    latency_v2_sd_full_nullvm_both_p50 = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p50')

    latency_none_mtls_base_p90 = get_latency_vs_qps_y_series(df, '_none_mtls_baseline', 'p90')
    latency_none_mtls_both_p90 = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p90')
    latency_none_plaintext_both_p90 = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p90')
    latency_v2_stats_nullvm_both_p90 = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p90')
    latency_v2_stats_wasm_both_p90 = get_latency_vs_qps_y_series(df, '_v2-stats-wasm_both', 'p90')
    latency_v2_sd_nologging_nullvm_both_p90 = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p90')
    latency_v2_sd_full_nullvm_both_p90 = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p90')

    latency_none_mtls_base_p99 = get_latency_vs_qps_y_series(df, '_none_mtls_baseline', 'p99')
    latency_none_mtls_both_p99 = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p99')
    latency_none_plaintext_both_p99 = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p99')
    latency_v2_stats_nullvm_both_p99 = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p99')
    latency_v2_stats_wasm_both_p99 = get_latency_vs_qps_y_series(df, '_v2-stats-wasm_both', 'p99')
    latency_v2_sd_nologging_nullvm_both_p99 = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99')
    latency_v2_sd_full_nullvm_both_p99 = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p99')

    latency_none_mtls_base_p999 = get_latency_vs_qps_y_series(df, '_none_mtls_baseline', 'p99.9')
    latency_none_mtls_both_p999 = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p99.9')
    latency_none_plaintext_both_p999 = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p99.9')
    latency_v2_stats_nullvm_both_p999 = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p99.9')
    latency_v2_stats_wasm_both_p999 = get_latency_vs_qps_y_series(df, '_v2-stats-wasm_both', 'p99.9')
    latency_v2_sd_nologging_nullvm_both_p999 = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99.9')
    latency_v2_sd_full_nullvm_both_p999 = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p99.9')

    context = {'latency_none_mtls_base_p50': latency_none_mtls_base_p50,
               'latency_none_mtls_both_p50': latency_none_mtls_both_p50,
               'latency_none_plaintext_both_p50': latency_none_plaintext_both_p50,
               'latency_v2_stats_nullvm_both_p50': latency_v2_stats_nullvm_both_p50,
               'latency_v2_stats_wasm_both_p50': latency_v2_stats_wasm_both_p50,
               'latency_v2_sd_nologging_nullvm_both_p50': latency_v2_sd_nologging_nullvm_both_p50,
               'latency_v2_sd_full_nullvm_both_p50': latency_v2_sd_full_nullvm_both_p50,
               'latency_none_mtls_base_p90': latency_none_mtls_base_p90,
               'latency_none_mtls_both_p90': latency_none_mtls_both_p90,
               'latency_none_plaintext_both_p90': latency_none_plaintext_both_p90,
               'latency_v2_stats_nullvm_both_p90': latency_v2_stats_nullvm_both_p90,
               'latency_v2_stats_wasm_both_p90': latency_v2_stats_wasm_both_p90,
               'latency_v2_sd_nologging_nullvm_both_p90': latency_v2_sd_nologging_nullvm_both_p90,
               'latency_v2_sd_full_nullvm_both_p90': latency_v2_sd_full_nullvm_both_p90,
               'latency_none_mtls_base_p99': latency_none_mtls_base_p99,
               'latency_none_mtls_both_p99': latency_none_mtls_both_p99,
               'latency_none_plaintext_both_p99': latency_none_plaintext_both_p99,
               'latency_v2_stats_nullvm_both_p99': latency_v2_stats_nullvm_both_p99,
               'latency_v2_stats_wasm_both_p99': latency_v2_stats_wasm_both_p99,
               'latency_v2_sd_nologging_nullvm_both_p99': latency_v2_sd_nologging_nullvm_both_p99,
               'latency_v2_sd_full_nullvm_both_p99': latency_v2_sd_full_nullvm_both_p99,
               'latency_none_mtls_base_p999': latency_none_mtls_base_p999,
               'latency_none_mtls_both_p999': latency_none_mtls_both_p999,
               'latency_none_plaintext_both_p999': latency_none_plaintext_both_p999,
               'latency_v2_stats_nullvm_both_p999': latency_v2_stats_nullvm_both_p999,
               'latency_v2_stats_wasm_both_p999': latency_v2_stats_wasm_both_p999,
               'latency_v2_sd_nologging_nullvm_both_p999': latency_v2_sd_nologging_nullvm_both_p999,
               'latency_v2_sd_full_nullvm_both_p999': latency_v2_sd_full_nullvm_both_p999,
               }
    return context


def get_cpu_vs_qps_context(df):
    cpu_client_none_mtls_base = get_cpu_vs_qps_y_series(df, '_none_mtls_baseline', cpu_client_metric_name)
    cpu_client_none_mtls_both = get_cpu_vs_qps_y_series(df, '_none_mtls_both', cpu_client_metric_name)
    cpu_client_none_plaintext_both = get_cpu_vs_qps_y_series(df, '_none_plaintext_both', cpu_client_metric_name)
    cpu_client_v2_stats_nullvm_both = get_cpu_vs_qps_y_series(df, '_v2-stats-nullvm_both', cpu_client_metric_name)
    cpu_client_v2_stats_wasm_both = get_cpu_vs_qps_y_series(df, '_v2-stats-wasm_both', cpu_client_metric_name)
    cpu_client_v2_sd_nologging_nullvm_both = get_cpu_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', cpu_client_metric_name)
    cpu_client_v2_sd_full_nullvm_both = get_cpu_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', cpu_client_metric_name)

    cpu_server_none_mtls_base = get_cpu_vs_qps_y_series(df, '_none_mtls_baseline', cpu_server_metric_name)
    cpu_server_none_mtls_both = get_cpu_vs_qps_y_series(df, '_none_mtls_both', cpu_server_metric_name)
    cpu_server_none_plaintext_both = get_cpu_vs_qps_y_series(df, '_none_plaintext_both', cpu_server_metric_name)
    cpu_server_v2_stats_nullvm_both = get_cpu_vs_qps_y_series(df, '_v2-stats-nullvm_both', cpu_server_metric_name)
    cpu_server_v2_stats_wasm_both = get_cpu_vs_qps_y_series(df, '_v2-stats-wasm_both', cpu_server_metric_name)
    cpu_server_v2_sd_nologging_nullvm_both = get_cpu_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', cpu_server_metric_name)
    cpu_server_v2_sd_full_nullvm_both = get_cpu_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', cpu_server_metric_name)

    context = {'cpu_client_none_mtls_base': cpu_client_none_mtls_base,
               'cpu_client_none_mtls_both': cpu_client_none_mtls_both,
               'cpu_client_none_plaintext_both': cpu_client_none_plaintext_both,
               'cpu_client_v2_stats_nullvm_both': cpu_client_v2_stats_nullvm_both,
               'cpu_client_v2_stats_wasm_both': cpu_client_v2_stats_wasm_both,
               'cpu_client_v2_sd_nologging_nullvm_both': cpu_client_v2_sd_nologging_nullvm_both,
               'cpu_client_v2_sd_full_nullvm_both': cpu_client_v2_sd_full_nullvm_both,
               'cpu_server_none_mtls_base': cpu_server_none_mtls_base,
               'cpu_server_none_mtls_both': cpu_server_none_mtls_both,
               'cpu_server_none_plaintext_both': cpu_server_none_plaintext_both,
               'cpu_server_v2_stats_nullvm_both': cpu_server_v2_stats_nullvm_both,
               'cpu_server_v2_stats_wasm_both': cpu_server_v2_stats_wasm_both,
               'cpu_server_v2_sd_nologging_nullvm_both': cpu_server_v2_sd_nologging_nullvm_both,
               'cpu_server_v2_sd_full_nullvm_both': cpu_server_v2_sd_full_nullvm_both,
               }
    return context


def get_cpu_vs_conn_context(df):
    cpu_client_none_mtls_base = get_cpu_vs_conn_y_series(df, '_none_mtls_baseline', cpu_client_metric_name)
    cpu_client_none_mtls_both = get_cpu_vs_conn_y_series(df, '_none_mtls_both', cpu_client_metric_name)
    cpu_client_none_plaintext_both = get_cpu_vs_conn_y_series(df, '_none_plaintext_both', cpu_client_metric_name)
    cpu_client_v2_stats_nullvm_both = get_cpu_vs_conn_y_series(df, '_v2-stats-nullvm_both', cpu_client_metric_name)
    cpu_client_v2_stats_wasm_both = get_cpu_vs_conn_y_series(df, '_v2-stats-wasm_both', cpu_client_metric_name)
    cpu_client_v2_sd_nologging_nullvm_both = get_cpu_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', cpu_client_metric_name)
    cpu_client_v2_sd_full_nullvm_both = get_cpu_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', cpu_client_metric_name)

    cpu_server_none_mtls_base = get_cpu_vs_conn_y_series(df, '_none_mtls_baseline', cpu_server_metric_name)
    cpu_server_none_mtls_both = get_cpu_vs_conn_y_series(df, '_none_mtls_both', cpu_server_metric_name)
    cpu_server_none_plaintext_both = get_cpu_vs_conn_y_series(df, '_none_plaintext_both', cpu_server_metric_name)
    cpu_server_v2_stats_nullvm_both = get_cpu_vs_conn_y_series(df, '_v2-stats-nullvm_both', cpu_server_metric_name)
    cpu_server_v2_stats_wasm_both = get_cpu_vs_conn_y_series(df, '_v2-stats-wasm_both', cpu_server_metric_name)
    cpu_server_v2_sd_nologging_nullvm_both = get_cpu_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', cpu_server_metric_name)
    cpu_server_v2_sd_full_nullvm_both = get_cpu_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', cpu_server_metric_name)

    context = {'cpu_client_none_mtls_base': cpu_client_none_mtls_base,
               'cpu_client_none_mtls_both': cpu_client_none_mtls_both,
               'cpu_client_none_plaintext_both': cpu_client_none_plaintext_both,
               'cpu_client_v2_stats_nullvm_both': cpu_client_v2_stats_nullvm_both,
               'cpu_client_v2_stats_wasm_both': cpu_client_v2_stats_wasm_both,
               'cpu_client_v2_sd_nologging_nullvm_both': cpu_client_v2_sd_nologging_nullvm_both,
               'cpu_client_v2_sd_full_nullvm_both': cpu_client_v2_sd_full_nullvm_both,
               'cpu_server_none_mtls_base': cpu_server_none_mtls_base,
               'cpu_server_none_mtls_both': cpu_server_none_mtls_both,
               'cpu_server_none_plaintext_both': cpu_server_none_plaintext_both,
               'cpu_server_v2_stats_nullvm_both': cpu_server_v2_stats_nullvm_both,
               'cpu_server_v2_stats_wasm_both': cpu_server_v2_stats_wasm_both,
               'cpu_server_v2_sd_nologging_nullvm_both': cpu_server_v2_sd_nologging_nullvm_both,
               'cpu_server_v2_sd_full_nullvm_both': cpu_server_v2_sd_full_nullvm_both,
               }
    return context


def get_mem_vs_qps_context(df):
    mem_client_none_mtls_base = get_mem_vs_qps_y_series(df, '_none_mtls_baseline', mem_client_metric_name)
    mem_client_none_mtls_both = get_mem_vs_qps_y_series(df, '_none_mtls_both', mem_client_metric_name)
    mem_client_none_plaintext_both = get_mem_vs_qps_y_series(df, '_none_plaintext_both', mem_client_metric_name)
    mem_client_v2_stats_nullvm_both = get_mem_vs_qps_y_series(df, '_v2-stats-nullvm_both', mem_client_metric_name)
    mem_client_v2_stats_wasm_both = get_mem_vs_qps_y_series(df, '_v2-stats-wasm_both', mem_client_metric_name)
    mem_client_v2_sd_nologging_nullvm_both = get_mem_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', mem_client_metric_name)
    mem_client_v2_sd_full_nullvm_both = get_mem_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', mem_client_metric_name)

    mem_server_none_mtls_base = get_mem_vs_qps_y_series(df, '_none_mtls_baseline', mem_server_metric_name)
    mem_server_none_mtls_both = get_mem_vs_qps_y_series(df, '_none_mtls_both', mem_server_metric_name)
    mem_server_none_plaintext_both = get_mem_vs_qps_y_series(df, '_none_plaintext_both', mem_server_metric_name)
    mem_server_v2_stats_nullvm_both = get_mem_vs_qps_y_series(df, '_v2-stats-nullvm_both', mem_server_metric_name)
    mem_server_v2_stats_wasm_both = get_mem_vs_qps_y_series(df, '_v2-stats-wasm_both', mem_server_metric_name)
    mem_server_v2_sd_nologging_nullvm_both = get_mem_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', mem_server_metric_name)
    mem_server_v2_sd_full_nullvm_both = get_mem_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', mem_server_metric_name)

    context = {'mem_client_none_mtls_base': mem_client_none_mtls_base,
               'mem_client_none_mtls_both': mem_client_none_mtls_both,
               'mem_client_none_plaintext_both': mem_client_none_plaintext_both,
               'mem_client_v2_stats_nullvm_both': mem_client_v2_stats_nullvm_both,
               'mem_client_v2_stats_wasm_both': mem_client_v2_stats_wasm_both,
               'mem_client_v2_sd_nologging_nullvm_both': mem_client_v2_sd_nologging_nullvm_both,
               'mem_client_v2_sd_full_nullvm_both': mem_client_v2_sd_full_nullvm_both,
               'mem_server_none_mtls_base': mem_server_none_mtls_base,
               'mem_server_none_mtls_both': mem_server_none_mtls_both,
               'mem_server_none_plaintext_both': mem_server_none_plaintext_both,
               'mem_server_v2_stats_nullvm_both': mem_server_v2_stats_nullvm_both,
               'mem_server_v2_stats_wasm_both': mem_server_v2_stats_wasm_both,
               'mem_server_v2_sd_nologging_nullvm_both': mem_server_v2_sd_nologging_nullvm_both,
               'mem_server_v2_sd_full_nullvm_both': mem_server_v2_sd_full_nullvm_both,
               }
    return context


def get_mem_vs_conn_context(df):
    mem_client_none_mtls_base = get_mem_vs_conn_y_series(df, '_none_mtls_baseline', mem_client_metric_name)
    mem_client_none_mtls_both = get_mem_vs_conn_y_series(df, '_none_mtls_both', mem_client_metric_name)
    mem_client_none_plaintext_both = get_mem_vs_conn_y_series(df, '_none_plaintext_both', mem_client_metric_name)
    mem_client_v2_stats_nullvm_both = get_mem_vs_conn_y_series(df, '_v2-stats-nullvm_both', mem_client_metric_name)
    mem_client_v2_stats_wasm_both = get_mem_vs_conn_y_series(df, '_v2-stats-wasm_both', mem_client_metric_name)
    mem_client_v2_sd_nologging_nullvm_both = get_mem_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', mem_client_metric_name)
    mem_client_v2_sd_full_nullvm_both = get_mem_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', mem_client_metric_name)

    mem_server_none_mtls_base = get_mem_vs_conn_y_series(df, '_none_mtls_baseline', mem_server_metric_name)
    mem_server_none_mtls_both = get_mem_vs_conn_y_series(df, '_none_mtls_both', mem_server_metric_name)
    mem_server_none_plaintext_both = get_mem_vs_conn_y_series(df, '_none_plaintext_both', mem_server_metric_name)
    mem_server_v2_stats_nullvm_both = get_mem_vs_conn_y_series(df, '_v2-stats-nullvm_both', mem_server_metric_name)
    mem_server_v2_stats_wasm_both = get_mem_vs_conn_y_series(df, '_v2-stats-wasm_both', mem_server_metric_name)
    mem_server_v2_sd_nologging_nullvm_both = get_mem_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', mem_server_metric_name)
    mem_server_v2_sd_full_nullvm_both = get_mem_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', mem_server_metric_name)

    context = {'mem_client_none_mtls_base': mem_client_none_mtls_base,
               'mem_client_none_mtls_both': mem_client_none_mtls_both,
               'mem_client_none_plaintext_both': mem_client_none_plaintext_both,
               'mem_client_v2_stats_nullvm_both': mem_client_v2_stats_nullvm_both,
               'mem_client_v2_stats_wasm_both': mem_client_v2_stats_wasm_both,
               'mem_client_v2_sd_nologging_nullvm_both': mem_client_v2_sd_nologging_nullvm_both,
               'mem_client_v2_sd_full_nullvm_both': mem_client_v2_sd_full_nullvm_both,
               'mem_server_none_mtls_base': mem_server_none_mtls_base,
               'mem_server_none_mtls_both': mem_server_none_mtls_both,
               'mem_server_none_plaintext_both': mem_server_none_plaintext_both,
               'mem_server_v2_stats_nullvm_both': mem_server_v2_stats_nullvm_both,
               'mem_server_v2_stats_wasm_both': mem_server_v2_stats_wasm_both,
               'mem_server_v2_sd_nologging_nullvm_both': mem_server_v2_sd_nologging_nullvm_both,
               'mem_server_v2_sd_full_nullvm_both': mem_server_v2_sd_full_nullvm_both,
               }
    return context


def flame_graph(request):
    cur_href_links, cur_release_names, cur_release_dates, master_href_links, master_release_names, master_release_dates = download.download_benchmark_csv(60)
    cur_release_bundle = get_flame_graph_release_bundle(cur_release_dates, cur_release_names, cur_href_links)
    master_release_bundle = get_flame_graph_release_bundle(master_release_dates, master_release_names, master_href_links)

    context = {'current_release': current_release,
               'cur_release_bundle': cur_release_bundle,
               'master_release_bundle': master_release_bundle}

    return render(request, "flame_graph.html", context=context)


def get_flame_graph_release_bundle(release_dates, release_names, href_links):
    release_bundle = [[]] * len(release_names)
    gcs_prefix = "https://gcsweb.istio.io/"
    for i in range(len(release_names)):
        release_bundle[i] = [0] * 3
        release_bundle[i][0] = release_dates[i]
        release_bundle[i][1] = release_names[i]
        release_bundle[i][2] = gcs_prefix + href_links[i] + "/flamegraphs/"
    return release_bundle


def micro_benchmarks(request):
    return render(request, "micro_benchmarks.html")


# Latency Helpers
def get_latency_vs_conn_y_series(df, telemetry_mode, quantiles):
    y_series_data = []
    for thread in [2, 4, 8, 16, 32, 64]:
        data = df.query('ActualQPS == 1000 and NumThreads == @thread and Labels.str.endswith(@telemetry_mode)')
        data_get_helper(data, y_series_data, quantiles)
    return y_series_data


def get_latency_vs_qps_y_series(df, telemetry_mode, quantiles):
    y_series_data = []
    for qps in [10, 100, 500, 1000, 2000, 3000]:
        data = df.query('ActualQPS == @qps and NumThreads == 16 and Labels.str.endswith(@telemetry_mode)')
        data_get_helper(data, y_series_data, quantiles)
    return y_series_data


# CPU Helpers
def get_cpu_vs_qps_y_series(df, telemetry_mode, cpu_metric_name):
    y_series_data = []
    for qps in [10, 100, 500, 1000, 2000, 3000]:
        data = df.query('ActualQPS == @qps and NumThreads == 16 and Labels.str.endswith(@telemetry_mode)')
        data_get_helper(data, y_series_data, cpu_metric_name)
    return y_series_data


def get_cpu_vs_conn_y_series(df, telemetry_mode, cpu_metric_name):
    y_series_data = []
    for thread in [2, 4, 8, 16, 32, 64]:
        data = df.query('ActualQPS == 1000 and NumThreads == @thread and Labels.str.endswith(@telemetry_mode)')
        data_get_helper(data, y_series_data, cpu_metric_name)
    return y_series_data


# Memory Helpers
def get_mem_vs_qps_y_series(df, telemetry_mode, mem_metric_name):
    y_series_data = []
    for qps in [10, 100, 200, 400, 800, 1000]:
        data = df.query('ActualQPS == @qps and NumThreads == 16 and Labels.str.endswith(@telemetry_mode)')
        data_get_helper(data, y_series_data, mem_metric_name)
    return y_series_data


def get_mem_vs_conn_y_series(df, telemetry_mode, mem_metric_name):
    y_series_data = []
    for thread in [2, 4, 8, 16, 32, 64]:
        data = df.query('ActualQPS == 1000 and NumThreads == @thread and Labels.str.endswith(@telemetry_mode)')
        data_get_helper(data, y_series_data, mem_metric_name)
    return y_series_data


def data_get_helper(data, y_series_data, metric_name):
    metric_data = data.get(metric_name)
    if metric_data is None or len(metric_data) == 0:
        y_series_data.append('null')
    else:
        if metric_name in ['cpu', 'mem']:
            y_series_data.append(data[metric_name].head(1).values[0])
        else:
            y_series_data.append(data[metric_name].head(1).values[0]/1000)

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
cur_selected_release = []
master_selected_release = []
cpu_cur_selected_release = []
cpu_master_selected_release = []
current_release = [os.getenv('CUR_RELEASE')]


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
        cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(60)

        if request.method == "POST" and 'current_release_name' in request.POST:
            cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_release_names[0] + ".csv")
        # Parse data for the current release
        if len(cur_selected_release) > 1:
            cur_selected_release.pop(0)
        if len(cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cur_selected_release[0] + ".csv")

        release_context = get_lantency_vs_conn_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_release_names[0] + ".csv")
        # Parse data for the current release
        if len(master_selected_release) > 1:
            master_selected_release.pop(0)
        if len(master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + master_selected_release[0] + ".csv")

        latency_mixer_both_p50_master = get_latency_vs_conn_y_series(df, '_mixer_both', 'p50')
        latency_none_mtls_base_p50_master = get_latency_vs_conn_y_series(df, '_none_mtls_base', 'p50')
        latency_none_mtls_both_p50_master = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p50')
        latency_none_plaintext_both_p50_master = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p50')
        latency_v2_stats_nullvm_both_p50_master = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p50')
        latency_v2_sd_nologging_nullvm_both_p50_master = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p50')
        latency_v2_sd_full_nullvm_both_p50_master = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p50')

        latency_mixer_both_p90_master = get_latency_vs_conn_y_series(df, '_mixer_both', 'p90')
        latency_none_mtls_base_p90_master = get_latency_vs_conn_y_series(df, '_none_mtls_base', 'p90')
        latency_none_mtls_both_p90_master = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p90')
        latency_none_plaintext_both_p90_master = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p90')
        latency_v2_stats_nullvm_both_p90_master = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p90')
        latency_v2_sd_nologging_nullvm_both_p90_master = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p90')
        latency_v2_sd_full_nullvm_both_p90_master = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p90')

        latency_mixer_both_p99_master = get_latency_vs_conn_y_series(df, '_mixer_both', 'p99')
        latency_none_mtls_base_p99_master = get_latency_vs_conn_y_series(df, '_none_mtls_base', 'p99')
        latency_none_mtls_both_p99_master = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p99')
        latency_none_plaintext_both_p99_master = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p99')
        latency_v2_stats_nullvm_both_p99_master = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p99')
        latency_v2_sd_nologging_nullvm_both_p99_master = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99')
        latency_v2_sd_full_nullvm_both_p99_master = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p99')

        other_context = {'current_release': current_release,
                         'cur_selected_release': cur_selected_release,
                         'master_selected_release':  master_selected_release,
                         'cur_release_names': cur_release_names,
                         'master_release_names': master_release_names,
                         }

        master_context = {'latency_mixer_both_p50_master': latency_mixer_both_p50_master,
                          'latency_none_mtls_base_p50_master': latency_none_mtls_base_p50_master,
                          'latency_none_mtls_both_p50_master': latency_none_mtls_both_p50_master,
                          'latency_none_plaintext_both_p50_master': latency_none_plaintext_both_p50_master,
                          'latency_v2_stats_nullvm_both_p50_master': latency_v2_stats_nullvm_both_p50_master,
                          'latency_v2_sd_nologging_nullvm_both_p50_master': latency_v2_sd_nologging_nullvm_both_p50_master,
                          'latency_v2_sd_full_nullvm_both_p50_master': latency_v2_sd_full_nullvm_both_p50_master,
                          'latency_mixer_both_p90_master': latency_mixer_both_p90_master,
                          'latency_none_mtls_base_p90_master': latency_none_mtls_base_p90_master,
                          'latency_none_mtls_both_p90_master': latency_none_mtls_both_p90_master,
                          'latency_none_plaintext_both_p90_master': latency_none_plaintext_both_p90_master,
                          'latency_v2_stats_nullvm_both_p90_master': latency_v2_stats_nullvm_both_p90_master,
                          'latency_v2_sd_nologging_nullvm_both_p90_master': latency_v2_sd_nologging_nullvm_both_p90_master,
                          'latency_v2_sd_full_nullvm_both_p90_master': latency_v2_sd_full_nullvm_both_p90_master,
                          'latency_mixer_both_p99_master': latency_mixer_both_p99_master,
                          'latency_none_mtls_base_p99_master': latency_none_mtls_base_p99_master,
                          'latency_none_mtls_both_p99_master': latency_none_mtls_both_p99_master,
                          'latency_none_plaintext_both_p99_master': latency_none_plaintext_both_p99_master,
                          'latency_v2_stats_nullvm_both_p99_master': latency_v2_stats_nullvm_both_p99_master,
                          'latency_v2_sd_nologging_nullvm_both_p99_master': latency_v2_sd_nologging_nullvm_both_p99_master,
                          'latency_v2_sd_full_nullvm_both_p99_master': latency_v2_sd_full_nullvm_both_p99_master,
                          }

        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))

        return render(request, "latency_vs_conn.html", context=context)


def latency_vs_qps(request, uploaded_csv_url=None):
    if uploaded_csv_url is not None:
        uploaded_csv_path = cwd + uploaded_csv_url
        df = pd.read_csv(uploaded_csv_path)
        context = get_lantency_vs_qps_context(df)
        os.remove(uploaded_csv_path)
        return context
    else:
        cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(60)

        if request.method == "POST" and 'current_release_name' in request.POST:
            cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_release_names[0] + ".csv")
        # Parse data for the current release
        if len(cur_selected_release) > 1:
            cur_selected_release.pop(0)
        if len(cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cur_selected_release[0] + ".csv")

        release_context = get_lantency_vs_qps_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_release_names[0] + ".csv")
        # Parse data for the current release
        if len(master_selected_release) > 1:
            master_selected_release.pop(0)
        if len(master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + master_selected_release[0] + ".csv")

        latency_mixer_both_p50_master = get_latency_vs_qps_y_series(df, '_mixer_both', 'p50')
        latency_none_mtls_base_p50_master = get_latency_vs_qps_y_series(df, '_none_mtls_base', 'p50')
        latency_none_mtls_both_p50_master = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p50')
        latency_none_plaintext_both_p50_master = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p50')
        latency_v2_stats_nullvm_both_p50_master = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p50')
        latency_v2_sd_nologging_nullvm_both_p50_master = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p50')
        latency_v2_sd_full_nullvm_both_p50_master = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p50')

        latency_mixer_both_p90_master = get_latency_vs_qps_y_series(df, '_mixer_both', 'p90')
        latency_none_mtls_base_p90_master = get_latency_vs_qps_y_series(df, '_none_mtls_base', 'p90')
        latency_none_mtls_both_p90_master = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p90')
        latency_none_plaintext_both_p90_master = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p90')
        latency_v2_stats_nullvm_both_p90_master = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p90')
        latency_v2_sd_nologging_nullvm_both_p90_master = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p90')
        latency_v2_sd_full_nullvm_both_p90_master = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p90')

        latency_mixer_both_p99_master = get_latency_vs_qps_y_series(df, '_mixer_both', 'p99')
        latency_none_mtls_base_p99_master = get_latency_vs_qps_y_series(df, '_none_mtls_base', 'p99')
        latency_none_mtls_both_p99_master = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p99')
        latency_none_plaintext_both_p99_master = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p99')
        latency_v2_stats_nullvm_both_p99_master = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p99')
        latency_v2_sd_nologging_nullvm_both_p99_master = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99')
        latency_v2_sd_full_nullvm_both_p99_master = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p99')

        other_context = {'current_release': current_release,
                         'cur_selected_release': cur_selected_release,
                         'master_selected_release':  master_selected_release,
                         'cur_release_names': cur_release_names,
                         'master_release_names': master_release_names,
                         }

        master_context = {'latency_mixer_both_p50_master': latency_mixer_both_p50_master,
                          'latency_none_mtls_base_p50_master': latency_none_mtls_base_p50_master,
                          'latency_none_mtls_both_p50_master': latency_none_mtls_both_p50_master,
                          'latency_none_plaintext_both_p50_master': latency_none_plaintext_both_p50_master,
                          'latency_v2_stats_nullvm_both_p50_master': latency_v2_stats_nullvm_both_p50_master,
                          'latency_v2_sd_nologging_nullvm_both_p50_master': latency_v2_sd_nologging_nullvm_both_p50_master,
                          'latency_v2_sd_full_nullvm_both_p50_master': latency_v2_sd_full_nullvm_both_p50_master,
                          'latency_mixer_both_p90_master': latency_mixer_both_p90_master,
                          'latency_none_mtls_base_p90_master': latency_none_mtls_base_p90_master,
                          'latency_none_mtls_both_p90_master': latency_none_mtls_both_p90_master,
                          'latency_none_plaintext_both_p90_master': latency_none_plaintext_both_p90_master,
                          'latency_v2_stats_nullvm_both_p90_master': latency_v2_stats_nullvm_both_p90_master,
                          'latency_v2_sd_nologging_nullvm_both_p90_master': latency_v2_sd_nologging_nullvm_both_p90_master,
                          'latency_v2_sd_full_nullvm_both_p90_master': latency_v2_sd_full_nullvm_both_p90_master,
                          'latency_mixer_both_p99_master': latency_mixer_both_p99_master,
                          'latency_none_mtls_base_p99_master': latency_none_mtls_base_p99_master,
                          'latency_none_mtls_both_p99_master': latency_none_mtls_both_p99_master,
                          'latency_none_plaintext_both_p99_master': latency_none_plaintext_both_p99_master,
                          'latency_v2_stats_nullvm_both_p99_master': latency_v2_stats_nullvm_both_p99_master,
                          'latency_v2_sd_nologging_nullvm_both_p99_master': latency_v2_sd_nologging_nullvm_both_p99_master,
                          'latency_v2_sd_full_nullvm_both_p99_master': latency_v2_sd_full_nullvm_both_p99_master,
                          }
        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))

        return render(request, "latency_vs_qps.html", context=context)


def cpu_memory(request, uploaded_csv_url=None):
    if uploaded_csv_url is not None:
        uploaded_csv_path = cwd + uploaded_csv_url
        df = pd.read_csv(uploaded_csv_path)
        context = get_cpu_mem_context(df)
        os.remove(uploaded_csv_path)
        return context
    else:
        cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(60)

        if request.method == "POST" and 'current_release_name' in request.POST:
            cpu_cur_selected_release.append(request.POST['current_release_name'])

        df = pd.read_csv(perf_data_path + "cur_temp.csv")

        if cur_release_names is not None and len(cur_release_names) > 0:
            df = pd.read_csv(perf_data_path + cur_release_names[0] + ".csv")
        # Parse data for the current release
        if len(cpu_cur_selected_release) > 1:
            cpu_cur_selected_release.pop(0)
        if len(cpu_cur_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cpu_cur_selected_release[0] + ".csv")

        release_context = get_cpu_mem_context(df)

        # Parse data for the master
        if request.method == "POST" and 'master_release_name' in request.POST:
            cpu_master_selected_release.append(request.POST['master_release_name'])

        df = pd.read_csv(perf_data_path + "master_temp.csv")

        if master_release_names is not None and len(master_release_names) > 0:
            df = pd.read_csv(perf_data_path + master_release_names[0] + ".csv")
        # Parse data for the current release
        if len(cpu_master_selected_release) > 1:
            cpu_master_selected_release.pop(0)
        if len(cpu_master_selected_release) > 0:
            df = pd.read_csv(perf_data_path + cpu_master_selected_release[0] + ".csv")

        cpu_mixer_both_master = get_cpu_y_series(df, '_mixer_both')
        cpu_none_mtls_base_master = get_cpu_y_series(df, '_none_mtls_base')
        cpu_none_mtls_both_master = get_cpu_y_series(df, '_none_mtls_both')
        cpu_none_plaintext_both_master = get_cpu_y_series(df, '_none_plaintext_both')
        cpu_v2_stats_nullvm_both_master = get_cpu_y_series(df, '_v2-stats-nullvm_both')
        cpu_v2_sd_nologging_nullvm_both_master = get_cpu_y_series(df, '_v2-sd-nologging-nullvm_both')
        cpu_v2_sd_full_nullvm_both_master = get_cpu_y_series(df, '_v2-sd-full-nullvm_both')

        mem_mixer_both_master = get_mem_y_series(df, '_mixer_both')
        mem_none_mtls_base_master = get_mem_y_series(df, '_none_mtls_base')
        mem_none_mtls_both_master = get_mem_y_series(df, '_none_mtls_both')
        mem_none_plaintext_both_master = get_mem_y_series(df, '_none_plaintext_both')
        mem_v2_stats_nullvm_both_master = get_mem_y_series(df, '_v2-stats-nullvm_both')
        mem_v2_sd_nologging_nullvm_both_master = get_mem_y_series(df, '_v2-sd-nologging-nullvm_both')
        mem_v2_sd_full_nullvm_both_master = get_mem_y_series(df, '_v2-sd-full-nullvm_both')

        other_context = {'current_release': current_release,
                         'cpu_cur_selected_release': cpu_cur_selected_release,
                         'cpu_master_selected_release': cpu_master_selected_release,
                         'cur_release_names': cur_release_names,
                         'master_release_names': master_release_names,
                         }

        master_context = {'cpu_mixer_both_master': cpu_mixer_both_master,
                          'cpu_none_mtls_base_master': cpu_none_mtls_base_master,
                          'cpu_none_mtls_both_master': cpu_none_mtls_both_master,
                          'cpu_none_plaintext_both_master': cpu_none_plaintext_both_master,
                          'cpu_v2_stats_nullvm_both_master': cpu_v2_stats_nullvm_both_master,
                          'cpu_v2_sd_nologging_nullvm_both_master': cpu_v2_sd_nologging_nullvm_both_master,
                          'cpu_v2_sd_full_nullvm_both_master': cpu_v2_sd_full_nullvm_both_master,
                          'mem_mixer_both_master': mem_mixer_both_master,
                          'mem_none_mtls_base_master': mem_none_mtls_base_master,
                          'mem_none_mtls_both_master': mem_none_mtls_both_master,
                          'mem_none_plaintext_both_master': mem_none_plaintext_both_master,
                          'mem_v2_stats_nullvm_both_master': mem_v2_stats_nullvm_both_master,
                          'mem_v2_sd_nologging_nullvm_both_master': mem_v2_sd_nologging_nullvm_both_master,
                          'mem_v2_sd_full_nullvm_both_master': mem_v2_sd_full_nullvm_both_master
                          }

        context = reduce(lambda x, y: dict(x, **y), (other_context, release_context, master_context))
        return render(request, "cpu_memory.html", context=context)


def get_lantency_vs_conn_context(df):
    latency_mixer_both_p50 = get_latency_vs_conn_y_series(df, '_mixer_both', 'p50')
    latency_none_mtls_base_p50 = get_latency_vs_conn_y_series(df, '_none_mtls_base', 'p50')
    latency_none_mtls_both_p50 = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p50')
    latency_none_plaintext_both_p50 = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p50')
    latency_v2_stats_nullvm_both_p50 = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p50')
    latency_v2_sd_nologging_nullvm_both_p50 = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p50')
    latency_v2_sd_full_nullvm_both_p50 = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p50')

    latency_mixer_both_p90 = get_latency_vs_conn_y_series(df, '_mixer_both', 'p90')
    latency_none_mtls_base_p90 = get_latency_vs_conn_y_series(df, '_none_mtls_base', 'p90')
    latency_none_mtls_both_p90 = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p90')
    latency_none_plaintext_both_p90 = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p90')
    latency_v2_stats_nullvm_both_p90 = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p90')
    latency_v2_sd_nologging_nullvm_both_p90 = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p90')
    latency_v2_sd_full_nullvm_both_p90 = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p90')

    latency_mixer_both_p99 = get_latency_vs_conn_y_series(df, '_mixer_both', 'p99')
    latency_none_mtls_base_p99 = get_latency_vs_conn_y_series(df, '_none_mtls_base', 'p99')
    latency_none_mtls_both_p99 = get_latency_vs_conn_y_series(df, '_none_mtls_both', 'p99')
    latency_none_plaintext_both_p99 = get_latency_vs_conn_y_series(df, '_none_plaintext_both', 'p99')
    latency_v2_stats_nullvm_both_p99 = get_latency_vs_conn_y_series(df, '_v2-stats-nullvm_both', 'p99')
    latency_v2_sd_nologging_nullvm_both_p99 = get_latency_vs_conn_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99')
    latency_v2_sd_full_nullvm_both_p99 = get_latency_vs_conn_y_series(df, '_v2-sd-full-nullvm_both', 'p99')

    context = {'latency_mixer_both_p50': latency_mixer_both_p50,
               'latency_none_mtls_base_p50': latency_none_mtls_base_p50,
               'latency_none_mtls_both_p50': latency_none_mtls_both_p50,
               'latency_none_plaintext_both_p50': latency_none_plaintext_both_p50,
               'latency_v2_stats_nullvm_both_p50': latency_v2_stats_nullvm_both_p50,
               'latency_v2_sd_nologging_nullvm_both_p50': latency_v2_sd_nologging_nullvm_both_p50,
               'latency_v2_sd_full_nullvm_both_p50': latency_v2_sd_full_nullvm_both_p50,
               'latency_mixer_both_p90': latency_mixer_both_p90,
               'latency_none_mtls_base_p90': latency_none_mtls_base_p90,
               'latency_none_mtls_both_p90': latency_none_mtls_both_p90,
               'latency_none_plaintext_both_p90': latency_none_plaintext_both_p90,
               'latency_v2_stats_nullvm_both_p90': latency_v2_stats_nullvm_both_p90,
               'latency_v2_sd_nologging_nullvm_both_p90': latency_v2_sd_nologging_nullvm_both_p90,
               'latency_v2_sd_full_nullvm_both_p90': latency_v2_sd_full_nullvm_both_p90,
               'latency_mixer_both_p99': latency_mixer_both_p99,
               'latency_none_mtls_base_p99': latency_none_mtls_base_p99,
               'latency_none_mtls_both_p99': latency_none_mtls_both_p99,
               'latency_none_plaintext_both_p99': latency_none_plaintext_both_p99,
               'latency_v2_stats_nullvm_both_p99': latency_v2_stats_nullvm_both_p99,
               'latency_v2_sd_nologging_nullvm_both_p99': latency_v2_sd_nologging_nullvm_both_p99,
               'latency_v2_sd_full_nullvm_both_p99': latency_v2_sd_full_nullvm_both_p99,
               }
    return context


def get_lantency_vs_qps_context(df):
    latency_mixer_both_p50 = get_latency_vs_qps_y_series(df, '_mixer_both', 'p50')
    latency_none_mtls_base_p50 = get_latency_vs_qps_y_series(df, '_none_mtls_base', 'p50')
    latency_none_mtls_both_p50 = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p50')
    latency_none_plaintext_both_p50 = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p50')
    latency_v2_stats_nullvm_both_p50 = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p50')
    latency_v2_sd_nologging_nullvm_both_p50 = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p50')
    latency_v2_sd_full_nullvm_both_p50 = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p50')

    latency_mixer_both_p90 = get_latency_vs_qps_y_series(df, '_mixer_both', 'p90')
    latency_none_mtls_base_p90 = get_latency_vs_qps_y_series(df, '_none_mtls_base', 'p90')
    latency_none_mtls_both_p90 = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p90')
    latency_none_plaintext_both_p90 = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p90')
    latency_v2_stats_nullvm_both_p90 = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p90')
    latency_v2_sd_nologging_nullvm_both_p90 = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p90')
    latency_v2_sd_full_nullvm_both_p90 = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p90')

    latency_mixer_both_p99 = get_latency_vs_qps_y_series(df, '_mixer_both', 'p99')
    latency_none_mtls_base_p99 = get_latency_vs_qps_y_series(df, '_none_mtls_base', 'p99')
    latency_none_mtls_both_p99 = get_latency_vs_qps_y_series(df, '_none_mtls_both', 'p99')
    latency_none_plaintext_both_p99 = get_latency_vs_qps_y_series(df, '_none_plaintext_both', 'p99')
    latency_v2_stats_nullvm_both_p99 = get_latency_vs_qps_y_series(df, '_v2-stats-nullvm_both', 'p99')
    latency_v2_sd_nologging_nullvm_both_p99 = get_latency_vs_qps_y_series(df, '_v2-sd-nologging-nullvm_both', 'p99')
    latency_v2_sd_full_nullvm_both_p99 = get_latency_vs_qps_y_series(df, '_v2-sd-full-nullvm_both', 'p99')

    context = {'latency_mixer_both_p50': latency_mixer_both_p50,
               'latency_none_mtls_base_p50': latency_none_mtls_base_p50,
               'latency_none_mtls_both_p50': latency_none_mtls_both_p50,
               'latency_none_plaintext_both_p50': latency_none_plaintext_both_p50,
               'latency_v2_stats_nullvm_both_p50': latency_v2_stats_nullvm_both_p50,
               'latency_v2_sd_nologging_nullvm_both_p50': latency_v2_sd_nologging_nullvm_both_p50,
               'latency_v2_sd_full_nullvm_both_p50': latency_v2_sd_full_nullvm_both_p50,
               'latency_mixer_both_p90': latency_mixer_both_p90,
               'latency_none_mtls_base_p90': latency_none_mtls_base_p90,
               'latency_none_mtls_both_p90': latency_none_mtls_both_p90,
               'latency_none_plaintext_both_p90': latency_none_plaintext_both_p90,
               'latency_v2_stats_nullvm_both_p90': latency_v2_stats_nullvm_both_p90,
               'latency_v2_sd_nologging_nullvm_both_p90': latency_v2_sd_nologging_nullvm_both_p90,
               'latency_v2_sd_full_nullvm_both_p90': latency_v2_sd_full_nullvm_both_p90,
               'latency_mixer_both_p99': latency_mixer_both_p99,
               'latency_none_mtls_base_p99': latency_none_mtls_base_p99,
               'latency_none_mtls_both_p99': latency_none_mtls_both_p99,
               'latency_none_plaintext_both_p99': latency_none_plaintext_both_p99,
               'latency_v2_stats_nullvm_both_p99': latency_v2_stats_nullvm_both_p99,
               'latency_v2_sd_nologging_nullvm_both_p99': latency_v2_sd_nologging_nullvm_both_p99,
               'latency_v2_sd_full_nullvm_both_p99': latency_v2_sd_full_nullvm_both_p99,
               }
    return context


def get_cpu_mem_context(df):
    cpu_mixer_both = get_cpu_y_series(df, '_mixer_both')
    cpu_none_mtls_base = get_cpu_y_series(df, '_none_mtls_base')
    cpu_none_mtls_both = get_cpu_y_series(df, '_none_mtls_both')
    cpu_none_plaintext_both = get_cpu_y_series(df, '_none_plaintext_both')
    cpu_v2_stats_nullvm_both = get_cpu_y_series(df, '_v2-stats-nullvm_both')
    cpu_v2_sd_nologging_nullvm_both = get_cpu_y_series(df, '_v2-sd-nologging-nullvm_both')
    cpu_v2_sd_full_nullvm_both = get_cpu_y_series(df, '_v2-sd-full-nullvm_both')

    mem_mixer_both = get_mem_y_series(df, '_mixer_both')
    mem_none_mtls_base = get_mem_y_series(df, '_none_mtls_base')
    mem_none_mtls_both = get_mem_y_series(df, '_none_mtls_both')
    mem_none_plaintext_both = get_mem_y_series(df, '_none_plaintext_both')
    mem_v2_stats_nullvm_both = get_mem_y_series(df, '_v2-stats-nullvm_both')
    mem_v2_sd_nologging_nullvm_both = get_mem_y_series(df, '_v2-sd-nologging-nullvm_both')
    mem_v2_sd_full_nullvm_both = get_mem_y_series(df, '_v2-sd-full-nullvm_both')

    context = {'cpu_mixer_both': cpu_mixer_both,
               'cpu_none_mtls_base': cpu_none_mtls_base,
               'cpu_none_mtls_both': cpu_none_mtls_both,
               'cpu_none_plaintext_both': cpu_none_plaintext_both,
               'cpu_v2_stats_nullvm_both': cpu_v2_stats_nullvm_both,
               'cpu_v2_sd_nologging_nullvm_both': cpu_v2_sd_nologging_nullvm_both,
               'cpu_v2_sd_full_nullvm_both': cpu_v2_sd_full_nullvm_both,
               'mem_mixer_both': mem_mixer_both,
               'mem_none_mtls_base': mem_none_mtls_base,
               'mem_none_mtls_both': mem_none_mtls_both,
               'mem_none_plaintext_both': mem_none_plaintext_both,
               'mem_v2_stats_nullvm_both': mem_v2_stats_nullvm_both,
               'mem_v2_sd_nologging_nullvm_both': mem_v2_sd_nologging_nullvm_both,
               'mem_v2_sd_full_nullvm_both': mem_v2_sd_full_nullvm_both,
               }
    return context


def flame_graph(request):
    return render(request, "flame_graph.html")


def micro_benchmarks(request):
    return render(request, "micro_benchmarks.html")


# Helpers
def get_latency_vs_conn_y_series(df, mixer_mode, quantiles):
    y_series_data = []
    if ("serveronly" in mixer_mode) or ("clientonly" in mixer_mode):
        return []
    for thread in [2, 4, 8, 16, 32, 64]:
        data = df.query('ActualQPS == 1000 and NumThreads == @thread and Labels.str.endswith(@mixer_mode)')
        if not data[quantiles].head().empty:
            y_series_data.append(data[quantiles].head(1).values[0]/1000)
        else:
            y_series_data.append('null')
    return y_series_data


def get_latency_vs_qps_y_series(df, mixer_mode, quantiles):
    y_series_data = []
    if ("serveronly" in mixer_mode) or ("clientonly" in mixer_mode):
        return []
    for qps in [10, 100, 500, 1000, 2000, 3000]:
        data = df.query('ActualQPS == @qps and NumThreads == 16 and Labels.str.endswith(@mixer_mode)')
        if not data[quantiles].head().empty:
            y_series_data.append(data[quantiles].head(1).values[0]/1000)
        else:
            y_series_data.append('null')
    return y_series_data


def get_cpu_y_series(df, mixer_mode):
    y_series_data = []
    if ("serveronly" in mixer_mode) or ("clientonly" in mixer_mode):
        return []
    cpu_metric = 'cpu_mili_avg_fortioserver_deployment_proxy'
    for qps in [10, 100, 500, 1000, 2000, 3000]:
        data = df.query('ActualQPS == @qps and NumThreads == 16 and Labels.str.endswith(@mixer_mode)')
        if not data[cpu_metric].head().empty:
            y_series_data.append(data[cpu_metric].head(1).values[0])
        else:
            y_series_data.append('null')
    return y_series_data


def get_mem_y_series(df, mixer_mode):
    y_series_data = []
    if ("serveronly" in mixer_mode) or ("clientonly" in mixer_mode):
        return []
    mem_metric = 'mem_MB_max_fortioserver_deployment_proxy'
    for qps in [10, 100, 500, 1000, 2000, 3000]:
        data = df.query('ActualQPS == @qps and NumThreads == 16 and Labels.str.endswith(@mixer_mode)')
        if not data[mem_metric].head().empty:
            y_series_data.append(data[mem_metric].head(1).values[0])
        else:
            y_series_data.append('null')
    return y_series_data

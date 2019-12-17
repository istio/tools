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


# Create your views here.
def latency(request):
    cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(20)

    if request.method == "POST" and 'current_release_name' in request.POST:
        cur_selected_release.append(request.POST['current_release_name'])
    df = pd.read_csv(perf_data_path + cur_release_names[0] + ".csv")
    # Parse data for the current release
    if len(cur_selected_release) > 1:
        cur_selected_release.pop(0)
    if len(cur_selected_release) > 0:
        df = pd.read_csv(perf_data_path + cur_selected_release[0] + ".csv")

    latency_mixer_base_p50 = get_latency_y_series(df, '_mixer_base', 'p50')
    latency_mixer_serveronly_p50 = get_latency_y_series(df, '_mixer_serveronly', 'p50')
    latency_mixer_both_p50 = get_latency_y_series(df, '_mixer_both', 'p50')
    latency_nomixer_serveronly_p50 = get_latency_y_series(df, '_nomixer_serveronly', 'p50')
    latency_nomixer_both_p50 = get_latency_y_series(df, '_nomixer_both', 'p50')
    latency_v2_serveronly_p50 = get_latency_y_series(df, 'nullvm_serveronly', 'p50')
    latency_v2_both_p50 = get_latency_y_series(df, 'nullvm_both', 'p50')

    latency_mixer_base_p90 = get_latency_y_series(df, '_mixer_base', 'p90')
    latency_mixer_serveronly_p90 = get_latency_y_series(df, '_mixer_serveronly', 'p90')
    latency_mixer_both_p90 = get_latency_y_series(df, '_mixer_both', 'p90')
    latency_nomixer_serveronly_p90 = get_latency_y_series(df, '_nomixer_serveronly', 'p90')
    latency_nomixer_both_p90 = get_latency_y_series(df, '_nomixer_both', 'p90')
    latency_v2_serveronly_p90 = get_latency_y_series(df, 'nullvm_serveronly', 'p90')
    latency_v2_both_p90 = get_latency_y_series(df, 'nullvm_both', 'p90')

    latency_mixer_base_p99 = get_latency_y_series(df, '_mixer_base', 'p99')
    latency_mixer_serveronly_p99 = get_latency_y_series(df, '_mixer_serveronly', 'p99')
    latency_mixer_both_p99 = get_latency_y_series(df, '_mixer_both', 'p99')
    latency_nomixer_serveronly_p99 = get_latency_y_series(df, '_nomixer_serveronly', 'p99')
    latency_nomixer_both_p99 = get_latency_y_series(df, '_nomixer_both', 'p99')
    latency_v2_serveronly_p99 = get_latency_y_series(df, 'nullvm_serveronly', 'p99')
    latency_v2_both_p99 = get_latency_y_series(df, 'nullvm_both', 'p99')

    # Parse data for the master
    if request.method == "POST" and 'master_release_name' in request.POST:
        master_selected_release.append(request.POST['master_release_name'])

    df = pd.read_csv(perf_data_path + master_release_names[0] + ".csv")
    # Parse data for the current release
    if len(master_selected_release) > 1:
        master_selected_release.pop(0)
    if len(master_selected_release) > 0:
        df = pd.read_csv(perf_data_path + master_selected_release[0] + ".csv")

    latency_mixer_base_p50_master = get_latency_y_series(df, '_mixer_base', 'p50')
    latency_mixer_serveronly_p50_master = get_latency_y_series(df, '_mixer_serveronly', 'p50')
    latency_mixer_both_p50_master = get_latency_y_series(df, '_mixer_both', 'p50')
    latency_nomixer_serveronly_p50_master = get_latency_y_series(df, '_nomixer_serveronly', 'p50')
    latency_nomixer_both_p50_master = get_latency_y_series(df, '_nomixer_both', 'p50')
    latency_v2_serveronly_p50_master = get_latency_y_series(df, 'nullvm_serveronly', 'p50')
    latency_v2_both_p50_master = get_latency_y_series(df, 'nullvm_both', 'p50')

    latency_mixer_base_p90_master = get_latency_y_series(df, '_mixer_base', 'p90')
    latency_mixer_serveronly_p90_master = get_latency_y_series(df, '_mixer_serveronly', 'p90')
    latency_mixer_both_p90_master = get_latency_y_series(df, '_mixer_both', 'p90')
    latency_nomixer_serveronly_p90_master = get_latency_y_series(df, '_nomixer_serveronly', 'p90')
    latency_nomixer_both_p90_master = get_latency_y_series(df, '_nomixer_both', 'p90')
    latency_v2_serveronly_p90_master = get_latency_y_series(df, 'nullvm_serveronly', 'p90')
    latency_v2_both_p90_master = get_latency_y_series(df, 'nullvm_both', 'p90')

    latency_mixer_base_p99_master = get_latency_y_series(df, '_mixer_base', 'p99')
    latency_mixer_serveronly_p99_master = get_latency_y_series(df, '_mixer_serveronly', 'p99')
    latency_mixer_both_p99_master = get_latency_y_series(df, '_mixer_both', 'p99')
    latency_nomixer_serveronly_p99_master = get_latency_y_series(df, '_nomixer_serveronly', 'p99')
    latency_nomixer_both_p99_master = get_latency_y_series(df, '_nomixer_both', 'p99')
    latency_v2_serveronly_p99_master = get_latency_y_series(df, 'nullvm_serveronly', 'p99')
    latency_v2_both_p99_master = get_latency_y_series(df, 'nullvm_both', 'p99')

    context = {'cur_selected_release': cur_selected_release,
               'master_selected_release':  master_selected_release,
               'cur_release_names': cur_release_names,
               'master_release_names': master_release_names,
               'latency_mixer_base_p50': latency_mixer_base_p50,
               'latency_mixer_serveronly_p50': latency_mixer_serveronly_p50,
               'latency_mixer_both_p50': latency_mixer_both_p50,
               'latency_nomixer_serveronly_p50': latency_nomixer_serveronly_p50,
               'latency_nomixer_both_p50': latency_nomixer_both_p50,
               'latency_v2_serveronly_p50': latency_v2_serveronly_p50,
               'latency_v2_both_p50': latency_v2_both_p50,
               'latency_mixer_base_p90': latency_mixer_base_p90,
               'latency_mixer_serveronly_p90': latency_mixer_serveronly_p90,
               'latency_mixer_both_p90': latency_mixer_both_p90,
               'latency_nomixer_serveronly_p90': latency_nomixer_serveronly_p90,
               'latency_nomixer_both_p90': latency_nomixer_both_p90,
               'latency_v2_serveronly_p90': latency_v2_serveronly_p90,
               'latency_v2_both_p90': latency_v2_both_p90,
               'latency_mixer_base_p99': latency_mixer_base_p99,
               'latency_mixer_serveronly_p99': latency_mixer_serveronly_p99,
               'latency_mixer_both_p99': latency_mixer_both_p99,
               'latency_nomixer_serveronly_p99': latency_nomixer_serveronly_p99,
               'latency_nomixer_both_p99': latency_nomixer_both_p99,
               'latency_v2_serveronly_p99': latency_v2_serveronly_p99,
               'latency_v2_both_p99': latency_v2_both_p99,
               'latency_mixer_base_p50_master': latency_mixer_base_p50_master,
               'latency_mixer_serveronly_p50_master': latency_mixer_serveronly_p50_master,
               'latency_mixer_both_p50_master': latency_mixer_both_p50_master,
               'latency_nomixer_serveronly_p50_master': latency_nomixer_serveronly_p50_master,
               'latency_nomixer_both_p50_master': latency_nomixer_both_p50_master,
               'latency_v2_serveronly_p50_master': latency_v2_serveronly_p50_master,
               'latency_v2_both_p50_master': latency_v2_both_p50_master,
               'latency_mixer_base_p90_master': latency_mixer_base_p90_master,
               'latency_mixer_serveronly_p90_master': latency_mixer_serveronly_p90_master,
               'latency_mixer_both_p90_master': latency_mixer_both_p90_master,
               'latency_nomixer_serveronly_p90_master': latency_nomixer_serveronly_p90_master,
               'latency_nomixer_both_p90_master': latency_nomixer_both_p90_master,
               'latency_v2_serveronly_p90_master': latency_v2_serveronly_p90_master,
               'latency_v2_both_p90_master': latency_v2_both_p90_master,
               'latency_mixer_base_p99_master': latency_mixer_base_p99_master,
               'latency_mixer_serveronly_p99_master': latency_mixer_serveronly_p99_master,
               'latency_mixer_both_p99_master': latency_mixer_both_p99_master,
               'latency_nomixer_serveronly_p99_master': latency_nomixer_serveronly_p99_master,
               'latency_nomixer_both_p99_master': latency_nomixer_both_p99_master,
               'latency_v2_serveronly_p99_master': latency_v2_serveronly_p99_master,
               'latency_v2_both_p99_master': latency_v2_both_p99_master,
               }
    return render(request, "latency.html", context=context)


def cpu_memory(request):
    cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(20)

    if request.method == "POST" and 'current_release_name' in request.POST:
        cpu_cur_selected_release.append(request.POST['current_release_name'])

    df = pd.read_csv(perf_data_path + cur_release_names[0] + ".csv")
    # Parse data for the current release
    if len(cpu_cur_selected_release) > 1:
        cpu_cur_selected_release.pop(0)
    if len(cpu_cur_selected_release) > 0:
        df = pd.read_csv(perf_data_path + cpu_cur_selected_release[0] + ".csv")

    cpu_mixer_base = get_cpu_y_series(df, '_mixer_base')
    cpu_mixer_serveronly = get_cpu_y_series(df, '_mixer_serveronly')
    cpu_mixer_both = get_cpu_y_series(df, '_mixer_both')
    cpu_nomixer_serveronly = get_cpu_y_series(df, '_nomixer_serveronly')
    cpu_nomixer_both = get_cpu_y_series(df, '_nomixer_both')
    cpu_v2_serveronly = get_cpu_y_series(df, 'nullvm_serveronly')
    cpu_v2_both = get_cpu_y_series(df, 'nullvm_both')

    mem_mixer_base = get_mem_y_series(df, '_mixer_base')
    mem_mixer_serveronly = get_mem_y_series(df, '_mixer_serveronly')
    mem_mixer_both = get_mem_y_series(df, '_mixer_both')
    mem_nomixer_serveronly = get_mem_y_series(df, '_nomixer_serveronly')
    mem_nomixer_both = get_mem_y_series(df, '_nomixer_both')
    mem_v2_serveronly = get_mem_y_series(df, 'nullvm_serveronly')
    mem_v2_both = get_mem_y_series(df, 'nullvm_both')

    # Parse data for the master
    if request.method == "POST" and 'master_release_name' in request.POST:
        cpu_master_selected_release.append(request.POST['master_release_name'])

    df = pd.read_csv(perf_data_path + master_release_names[0] + ".csv")
    # Parse data for the current release
    if len(cpu_master_selected_release) > 1:
        cpu_master_selected_release.pop(0)
    if len(cpu_master_selected_release) > 0:
        df = pd.read_csv(perf_data_path + cpu_master_selected_release[0] + ".csv")

    cpu_mixer_base_master = get_cpu_y_series(df, '_mixer_base')
    cpu_mixer_serveronly_master = get_cpu_y_series(df, '_mixer_serveronly')
    cpu_mixer_both_master = get_cpu_y_series(df, '_mixer_both')
    cpu_nomixer_serveronly_master = get_cpu_y_series(df, '_nomixer_serveronly')
    cpu_nomixer_both_master = get_cpu_y_series(df, '_nomixer_both')
    cpu_v2_serveronly_master = get_cpu_y_series(df, 'nullvm_serveronly')
    cpu_v2_both_master = get_cpu_y_series(df, 'nullvm_both')

    mem_mixer_base_master = get_mem_y_series(df, '_mixer_base')
    mem_mixer_serveronly_master = get_mem_y_series(df, '_mixer_serveronly')
    mem_mixer_both_master = get_mem_y_series(df, '_mixer_both')
    mem_nomixer_serveronly_master = get_mem_y_series(df, '_nomixer_serveronly')
    mem_nomixer_both_master = get_mem_y_series(df, '_nomixer_both')
    mem_v2_serveronly_master = get_mem_y_series(df, 'nullvm_serveronly')
    mem_v2_both_master = get_mem_y_series(df, 'nullvm_both')

    context = {'cpu_cur_selected_release': cpu_cur_selected_release,
               'cpu_master_selected_release':  cpu_master_selected_release,
               'cur_release_names': cur_release_names,
               'master_release_names': master_release_names,
               'cpu_mixer_base': cpu_mixer_base,
               'cpu_mixer_serveronly': cpu_mixer_serveronly,
               'cpu_mixer_both': cpu_mixer_both,
               'cpu_nomixer_serveronly': cpu_nomixer_serveronly,
               'cpu_nomixer_both': cpu_nomixer_both,
               'cpu_v2_serveronly': cpu_v2_serveronly,
               'cpu_v2_both': cpu_v2_both,
               'cpu_mixer_base_master': cpu_mixer_base_master,
               'cpu_mixer_serveronly_master': cpu_mixer_serveronly_master,
               'cpu_mixer_both_master': cpu_mixer_both_master,
               'cpu_nomixer_serveronly_master': cpu_nomixer_serveronly_master,
               'cpu_nomixer_both_master': cpu_nomixer_both_master,
               'cpu_v2_serveronly_master': cpu_v2_serveronly_master,
               'cpu_v2_both_master': cpu_v2_both_master,
               'mem_mixer_base': mem_mixer_base,
               'mem_mixer_serveronly': mem_mixer_serveronly,
               'mem_mixer_both': mem_mixer_both,
               'mem_nomixer_serveronly': mem_nomixer_serveronly,
               'mem_nomixer_both': mem_nomixer_both,
               'mem_v2_serveronly': mem_v2_serveronly,
               'mem_v2_both': mem_v2_both,
               'mem_mixer_base_master': mem_mixer_base_master,
               'mem_mixer_serveronly_master': mem_mixer_serveronly_master,
               'mem_mixer_both_master': mem_mixer_both_master,
               'mem_nomixer_serveronly_master': mem_nomixer_serveronly_master,
               'mem_nomixer_both_master': mem_nomixer_both_master,
               'mem_v2_serveronly_master': mem_v2_serveronly_master,
               'mem_v2_both_master': mem_v2_both_master,
               }
    return render(request, "cpu-memory.html", context=context)


def flame_graph(request):
    return render(request, "flame_graph.html")


def micro_benchmarks(request):
    return render(request, "micro_benchmarks.html")


# Helpers
def get_latency_y_series(df, mixer_mode, quantiles):
    y_series_data = []
    for thread in [2, 4, 8, 16, 32, 64]:
        data = df.query('ActualQPS == 1000 and NumThreads == @thread and Labels.str.endswith(@mixer_mode)')
        if not data[quantiles].head().empty:
            y_series_data.append(data[quantiles].head(1).values[0])
        else:
            y_series_data.append('null')
    return y_series_data


def get_cpu_y_series(df, mixer_mode):
    y_series_data = []
    cpu_metric = 'cpu_mili_avg_fortioserver_deployment_proxy'
    for qps in [10, 100, 500, 1000, 2000, 3000]:
        data = df.query('ActualQPS == @qps and NumThreads == 16  and Labels.str.endswith(@mixer_mode)')
        if not data[cpu_metric].head().empty:
            y_series_data.append(data[cpu_metric].head(1).values[0])
        else:
            y_series_data.append('null')
    return y_series_data


def get_mem_y_series(df, mixer_mode):
    y_series_data = []
    mem_metric = 'mem_MB_max_fortioserver_deployment_proxy'
    for qps in [10, 100, 500, 1000, 2000, 3000]:
        data = df.query('ActualQPS == @qps and NumThreads == 16  and Labels.str.endswith(@mixer_mode)')
        if not data[mem_metric].head().empty:
            y_series_data.append(data[mem_metric].head(1).values[0])
        else:
            y_series_data.append('null')
    return y_series_data

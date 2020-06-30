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
from helpers import download
import pandas as pd
import os


cwd = os.getcwd()
perf_data_path = cwd + "/perf_data/"
current_release = [os.getenv('CUR_RELEASE')]


# Create your views here.
def cur_pattern(request):
    cur_release_names, cur_release_dates, _, _ = download.download_benchmark_csv(40)

    none_mtls_base_p90_pattern = get_telemetry_mode_y_series(cur_release_names, cur_release_dates, '_none_mtls_base', 'p90')
    none_mtls_both_p90_pattern = get_telemetry_mode_y_series(cur_release_names, cur_release_dates, '_none_mtls_both', 'p90')
    v2_sd_full_nullvm_both_p90_pattern = get_telemetry_mode_y_series(cur_release_names, cur_release_dates, '_v2-sd-full-nullvm_both', 'p90')

    none_mtls_base_p99_pattern = get_telemetry_mode_y_series(cur_release_names, cur_release_dates, '_none_mtls_base', 'p99')
    none_mtls_both_p99_pattern = get_telemetry_mode_y_series(cur_release_names, cur_release_dates, '_none_mtls_both', 'p99')
    v2_sd_full_nullvm_both_p99_pattern = get_telemetry_mode_y_series(cur_release_names, cur_release_dates, '_v2-sd-full-nullvm_both', 'p99')

    context = {'current_release': current_release,
               'none_mtls_base_p90_pattern': none_mtls_base_p90_pattern,
               'none_mtls_both_p90_pattern': none_mtls_both_p90_pattern,
               'v2_sd_full_nullvm_both_p90_pattern': v2_sd_full_nullvm_both_p90_pattern,
               'none_mtls_base_p99_pattern': none_mtls_base_p99_pattern,
               'none_mtls_both_p99_pattern': none_mtls_both_p99_pattern,
               'v2_sd_full_nullvm_both_p99_pattern': v2_sd_full_nullvm_both_p99_pattern,
               }
    return render(request, "cur_pattern.html", context=context)


# Create your views here.
def master_pattern(request):
    _, _, master_release_names, master_release_dates = download.download_benchmark_csv(40)

    none_mtls_base_p90_pattern_master = get_telemetry_mode_y_series(master_release_names, master_release_dates, '_none_mtls_base', 'p90')
    none_mtls_both_p90_pattern_master = get_telemetry_mode_y_series(master_release_names, master_release_dates, '_none_mtls_both', 'p90')
    v2_sd_full_nullvm_both_p90_pattern_master = get_telemetry_mode_y_series(master_release_names, master_release_dates, '_v2-sd-full-nullvm_both', 'p90')

    none_mtls_base_p99_pattern_master = get_telemetry_mode_y_series(master_release_names, master_release_dates, '_none_mtls_base', 'p99')
    none_mtls_both_p99_pattern_master = get_telemetry_mode_y_series(master_release_names, master_release_dates, '_none_mtls_both', 'p99')
    v2_sd_full_nullvm_both_p99_pattern_master = get_telemetry_mode_y_series(master_release_names, master_release_dates, '_v2-sd-full-nullvm_both', 'p99')

    context = {'none_mtls_base_p90_pattern_master': none_mtls_base_p90_pattern_master,
               'none_mtls_both_p90_pattern_master': none_mtls_both_p90_pattern_master,
               'v2_sd_full_nullvm_both_p90_pattern_master': v2_sd_full_nullvm_both_p90_pattern_master,
               'none_mtls_base_p99_pattern_master': none_mtls_base_p99_pattern_master,
               'none_mtls_both_p99_pattern_master': none_mtls_both_p99_pattern_master,
               'v2_sd_full_nullvm_both_p99_pattern_master': v2_sd_full_nullvm_both_p99_pattern_master,
               }
    return render(request, "master_pattern.html", context=context)


# Helpers
def get_latency_y_data_point(df, telemetry_mode, quantiles):
    y_series_data = []
    data = df.query('ActualQPS == 1000 and NumThreads == 16 and Labels.str.endswith(@telemetry_mode)')
    if not data[quantiles].head().empty:
        y_series_data.append(data[quantiles].head(1).values[0] / 1000)
    else:
        y_series_data.append('null')
    return y_series_data


def get_telemetry_mode_y_series(release_names, release_dates, telemetry_mode, quantiles):
    pattern_data = [[]] * len(release_names)
    for i in range(len(release_names)):
        try:
            df = pd.read_csv(perf_data_path + release_names[i] + ".csv")
        except Exception as e:
            print(e)
            pattern_data[i] = release_dates[i] + ["null"]
        else:
            pattern_data[i] = release_dates[i] + get_latency_y_data_point(df, telemetry_mode, quantiles)
    return pattern_data

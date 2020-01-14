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
from helpers import outlier_detection as od
import pandas as pd
import os


cwd = os.getcwd()
perf_data_path = cwd + "/perf_data/"


# Create your views here.
def cur_alert(request):
    cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(40)

    cur_pattern_mixer_base_p90 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_base', 'p90')
    cur_pattern_mixer_serveronly_p90 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_serveronly', 'p90')
    cur_pattern_mixer_both_p90 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_both', 'p90')
    cur_pattern_none_serveronly_p90 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_none_serveronly', 'p90')
    cur_pattern_none_both_p90 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_none_both', 'p90')
    cur_pattern_v2_serveronly_p90 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, 'nullvm_serveronly', 'p90')
    cur_pattern_v2_both_p90 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, 'nullvm_both', 'p90')

    outliers_cur_mixer_serveronly_p90 = od.format_outliers(od.find_outliers(cur_pattern_mixer_serveronly_p90), 'mixer_serveronly')
    outliers_cur_mixer_both_p90 = od.format_outliers(od.find_outliers(cur_pattern_mixer_both_p90), 'mixer_both')
    outliers_cur_none_serveronly_p90 = od.format_outliers(od.find_outliers(cur_pattern_none_serveronly_p90), 'none_serveronly')
    outliers_cur_none_both_p90 = od.format_outliers(od.find_outliers(cur_pattern_none_both_p90), 'none_both')
    outliers_cur_v2_serveronly_p90 = od.format_outliers(od.find_outliers(cur_pattern_v2_serveronly_p90), 'v2_serveronly')
    outliers_cur_v2_both_p90 = od.format_outliers(od.find_outliers(cur_pattern_v2_both_p90), 'v2_both')

    outliers_cur_p90 = outliers_cur_mixer_serveronly_p90 + outliers_cur_mixer_both_p90 + \
        outliers_cur_none_serveronly_p90 + outliers_cur_none_both_p90 + \
        outliers_cur_v2_serveronly_p90 + outliers_cur_v2_both_p90

    cur_pattern_mixer_base_p99 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_base', 'p99')
    cur_pattern_mixer_serveronly_p99 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_serveronly', 'p99')
    cur_pattern_mixer_both_p99 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_both', 'p99')
    cur_pattern_none_serveronly_p99 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_none_serveronly', 'p99')
    cur_pattern_none_both_p99 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_none_both', 'p99')
    cur_pattern_v2_serveronly_p99 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, 'nullvm_serveronly', 'p99')
    cur_pattern_v2_both_p99 = get_mixer_mode_y_series(cur_release_names, cur_release_dates, 'nullvm_both', 'p99')

    outliers_cur_mixer_serveronly_p99 = od.format_outliers(od.find_outliers(cur_pattern_mixer_serveronly_p99), 'mixer_serveronly')
    outliers_cur_mixer_both_p99 = od.format_outliers(od.find_outliers(cur_pattern_mixer_both_p99), 'mixer_both')
    outliers_cur_none_serveronly_p99 = od.format_outliers(od.find_outliers(cur_pattern_none_serveronly_p99), 'none_serveronly')
    outliers_cur_none_both_p99 = od.format_outliers(od.find_outliers(cur_pattern_none_both_p99), 'none_both')
    outliers_cur_v2_serveronly_p99 = od.format_outliers(od.find_outliers(cur_pattern_v2_serveronly_p99), 'v2_serveronly')
    outliers_cur_v2_both_p99 = od.format_outliers(od.find_outliers(cur_pattern_v2_both_p99), 'v2_both')
    outliers_cur_p99 = outliers_cur_mixer_serveronly_p99 + outliers_cur_mixer_both_p99 + \
        outliers_cur_none_serveronly_p99 + outliers_cur_none_both_p99 + \
        outliers_cur_v2_serveronly_p99 + outliers_cur_v2_both_p99

    context = {'outliers_cur_p90':  outliers_cur_p90,
               'outliers_cur_p99': outliers_cur_p99,
               'cur_pattern_mixer_base_p90': cur_pattern_mixer_base_p90,
               'cur_pattern_mixer_serveronly_p90': cur_pattern_mixer_serveronly_p90,
               'cur_pattern_mixer_both_p90': cur_pattern_mixer_both_p90,
               'cur_pattern_none_serveronly_p90': cur_pattern_none_serveronly_p90,
               'cur_pattern_none_both_p90': cur_pattern_none_both_p90,
               'cur_pattern_v2_serveronly_p90': cur_pattern_v2_serveronly_p90,
               'cur_pattern_v2_both_p90': cur_pattern_v2_both_p90,
               'cur_pattern_mixer_base_p99': cur_pattern_mixer_base_p99,
               'cur_pattern_mixer_serveronly_p99': cur_pattern_mixer_serveronly_p99,
               'cur_pattern_mixer_both_p99': cur_pattern_mixer_both_p99,
               'cur_pattern_none_serveronly_p99': cur_pattern_none_serveronly_p99,
               'cur_pattern_none_both_p99': cur_pattern_none_both_p99,
               'cur_pattern_v2_serveronly_p99': cur_pattern_v2_serveronly_p99,
               'cur_pattern_v2_both_p99': cur_pattern_v2_both_p99
               }
    return render(request, "cur_alert.html", context=context)


# Create your views here.
def master_alert(request):
    cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(40)

    master_pattern_mixer_base_p90 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_base', 'p90')
    master_pattern_mixer_serveronly_p90 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_serveronly', 'p90')
    master_pattern_mixer_both_p90 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_both', 'p90')
    master_pattern_none_serveronly_p90 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_none_serveronly', 'p90')
    master_pattern_none_both_p90 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_none_both', 'p90')
    master_pattern_v2_serveronly_p90 = get_mixer_mode_y_series(master_release_names, master_release_dates, 'nullvm_serveronly', 'p90')
    master_pattern_v2_both_p90 = get_mixer_mode_y_series(master_release_names, master_release_dates, 'nullvm_both', 'p90')

    outliers_master_mixer_serveronly_p90 = od.format_outliers(od.find_outliers(master_pattern_mixer_serveronly_p90), 'mixer_serveronly')
    outliers_master_mixer_both_p90 = od.format_outliers(od.find_outliers(master_pattern_mixer_both_p90), 'mixer_both')
    outliers_master_none_serveronly_p90 = od.format_outliers(od.find_outliers(master_pattern_none_serveronly_p90), 'none_serveronly')
    outliers_master_none_both_p90 = od.format_outliers(od.find_outliers(master_pattern_none_both_p90), 'none_both')
    outliers_master_v2_serveronly_p90 = od.format_outliers(od.find_outliers(master_pattern_v2_serveronly_p90), 'v2_serveronly')
    outliers_master_v2_both_p90 = od.format_outliers(od.find_outliers(master_pattern_v2_both_p90), 'v2_both')
    outliers_master_p90 = outliers_master_mixer_serveronly_p90 + outliers_master_mixer_both_p90 + \
        outliers_master_none_serveronly_p90 + outliers_master_none_both_p90 + \
        outliers_master_v2_serveronly_p90 + outliers_master_v2_both_p90

    master_pattern_mixer_base_p99 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_base', 'p99')
    master_pattern_mixer_serveronly_p99 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_serveronly', 'p99')
    master_pattern_mixer_both_p99 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_both', 'p99')
    master_pattern_none_serveronly_p99 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_none_serveronly', 'p99')
    master_pattern_none_both_p99 = get_mixer_mode_y_series(master_release_names, master_release_dates, '_none_both', 'p99')
    master_pattern_v2_serveronly_p99 = get_mixer_mode_y_series(master_release_names, master_release_dates, 'nullvm_serveronly', 'p99')
    master_pattern_v2_both_p99 = get_mixer_mode_y_series(master_release_names, master_release_dates, 'nullvm_both', 'p99')

    outliers_master_mixer_serveronly_p99 = od.format_outliers(od.find_outliers(master_pattern_mixer_serveronly_p99), 'mixer_serveronly')
    outliers_master_mixer_both_p99 = od.format_outliers(od.find_outliers(master_pattern_mixer_both_p99), 'mixer_both')
    outliers_master_none_serveronly_p99 = od.format_outliers(od.find_outliers(master_pattern_none_serveronly_p99), 'none_serveronly')
    outliers_master_none_both_p99 = od.format_outliers(od.find_outliers(master_pattern_none_both_p99), 'none_both')
    outliers_master_v2_serveronly_p99 = od.format_outliers(od.find_outliers(master_pattern_v2_serveronly_p99), 'v2_serveronly')
    outliers_master_v2_both_p99 = od.format_outliers(od.find_outliers(master_pattern_v2_both_p99), 'v2_both')
    outliers_master_p99 = outliers_master_mixer_serveronly_p99 + outliers_master_mixer_both_p99 + \
        outliers_master_none_serveronly_p99 + outliers_master_none_both_p99 + \
        outliers_master_v2_serveronly_p99 + outliers_master_v2_both_p99

    context = {'outliers_master_p90':  outliers_master_p90,
               'outliers_master_p99': outliers_master_p99,
               'master_pattern_mixer_base_p90': master_pattern_mixer_base_p90,
               'master_pattern_mixer_serveronly_p90': master_pattern_mixer_serveronly_p90,
               'master_pattern_mixer_both_p90': master_pattern_mixer_both_p90,
               'master_pattern_none_serveronly_p90': master_pattern_none_serveronly_p90,
               'master_pattern_none_both_p90': master_pattern_none_both_p90,
               'master_pattern_v2_serveronly_p90': master_pattern_v2_serveronly_p90,
               'master_pattern_v2_both_p90': master_pattern_v2_both_p90,
               'master_pattern_mixer_base_p99': master_pattern_mixer_base_p99,
               'master_pattern_mixer_serveronly_p99': master_pattern_mixer_serveronly_p99,
               'master_pattern_mixer_both_p99': master_pattern_mixer_both_p99,
               'master_pattern_none_serveronly_p99': master_pattern_none_serveronly_p99,
               'master_pattern_none_both_p99': master_pattern_none_both_p99,
               'master_pattern_v2_serveronly_p99': master_pattern_v2_serveronly_p99,
               'master_pattern_v2_both_p99': master_pattern_v2_both_p99
               }
    return render(request, "master_alert.html", context=context)


# Helpers
def get_latency_y_data_point(df, mixer_mode, quantiles):
    y_series_data = []
    data = df.query('ActualQPS == 1000 and NumThreads == 16 and Labels.str.endswith(@mixer_mode)')
    if not data[quantiles].head().empty:
        y_series_data.append(data[quantiles].head(1).values[0]/1000)
    else:
        y_series_data.append('null')
    return y_series_data


def get_mixer_mode_y_series(release_names, release_dates, mixer_mode, quantiles):
    pattern_data = [[]] * len(release_names)
    for i in range(len(release_names)):
        try:
            df = pd.read_csv(perf_data_path + release_names[i] + ".csv")
        except Exception as e:
            print(e)
            pattern_data[i] = release_dates[i][:3] + ["null"] + [release_dates[i][3]] + [release_names[i]]
        else:
            pattern_data[i] = release_dates[i][:3] + get_latency_y_data_point(df, mixer_mode, quantiles) + \
                [release_dates[i][3]] + [release_names[i]]
    """
    The patten_data we get here is an array of array, each element is in the following format:
    ['2019', '12', '23', 4.478, '20191223', 'release-1.4.20191223-16.da6d4af0a2e1c3207edfd97f09d07a638c59e89a']
    """
    return pattern_data

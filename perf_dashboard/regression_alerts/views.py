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


# Create your views here.
def cur_alert(request):
    cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(40)

    cur_pattern_mixer_base = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_base')
    cur_pattern_mixer_serveronly = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_serveronly')
    cur_pattern_mixer_both = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_mixer_both')
    cur_pattern_nomixer_serveronly = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_nomixer_serveronly')
    cur_pattern_nomixer_both = get_mixer_mode_y_series(cur_release_names, cur_release_dates, '_nomixer_both')
    cur_pattern_v2_serveronly = get_mixer_mode_y_series(cur_release_names, cur_release_dates, 'nullvm_serveronly')
    cur_pattern_v2_both = get_mixer_mode_y_series(cur_release_names, cur_release_dates, 'nullvm_both')
    print("+++++++")
    print(cur_pattern_v2_serveronly)
    context = {'cur_pattern_mixer_base': cur_pattern_mixer_base,
               'cur_pattern_mixer_serveronly': cur_pattern_mixer_serveronly,
               'cur_pattern_mixer_both': cur_pattern_mixer_both,
               'cur_pattern_nomixer_serveronly': cur_pattern_nomixer_serveronly,
               'cur_pattern_nomixer_both': cur_pattern_nomixer_both,
               'cur_pattern_v2_serveronly': cur_pattern_v2_serveronly,
               'cur_pattern_v2_both': cur_pattern_v2_both}
    return render(request, "cur_alert.html", context=context)


# Create your views here.
def master_alert(request):
    cur_release_names, cur_release_dates, master_release_names, master_release_dates = download.download_benchmark_csv(40)

    master_pattern_mixer_base = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_base')
    master_pattern_mixer_serveronly = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_serveronly')
    master_pattern_mixer_both = get_mixer_mode_y_series(master_release_names, master_release_dates, '_mixer_both')
    master_pattern_nomixer_serveronly = get_mixer_mode_y_series(master_release_names, master_release_dates, '_nomixer_serveronly')
    master_pattern_nomixer_both = get_mixer_mode_y_series(master_release_names, master_release_dates, '_nomixer_both')
    master_pattern_v2_serveronly = get_mixer_mode_y_series(master_release_names, master_release_dates, 'nullvm_serveronly')
    master_pattern_v2_both = get_mixer_mode_y_series(master_release_names, master_release_dates, 'nullvm_both')

    context = {'master_pattern_mixer_base': master_pattern_mixer_base,
               'master_pattern_mixer_serveronly': master_pattern_mixer_serveronly,
               'master_pattern_mixer_both': master_pattern_mixer_both,
               'master_pattern_nomixer_serveronly': master_pattern_nomixer_serveronly,
               'master_pattern_nomixer_both': master_pattern_nomixer_both,
               'master_pattern_v2_serveronly': master_pattern_v2_serveronly,
               'master_pattern_v2_both': master_pattern_v2_both}
    return render(request, "master_alert.html", context=context)


# Helpers
def get_latency_y_data_point(df, mixer_mode):
    quantiles = 'p90'
    y_series_data = []
    data = df.query('ActualQPS == 1000 and NumThreads == 16 and Labels.str.endswith(@mixer_mode)')
    if not data[quantiles].head().empty:
        y_series_data.append(data[quantiles].head(1).values[0]/1000)
    else:
        y_series_data.append('null')
    return y_series_data


def get_mixer_mode_y_series(release_names, release_dates, mixer_mode):
    pattern_data = [[]] * len(release_names)
    for i in range(len(release_names)):
        try:
            df = pd.read_csv(perf_data_path + release_names[i] + ".csv")
        except Exception as e:
            print(e)
            pattern_data[i] = release_dates[i] + ["null"]
        else:
            pattern_data[i] = release_dates[i] + get_latency_y_data_point(df, mixer_mode)
    return pattern_data

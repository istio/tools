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
def monitoring_overview(request):
    return render(request, "monitoring_overview.html")


def cur_regression(request):
    cur_href_links, _, cur_release_dates, _, _, _ = download.download_benchmark_csv(60)

    latency_none_mtls_base_p90 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_none_mtls_baseline', 'p90')
    latency_none_mtls_both_p90 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_none_mtls_both', 'p90')
    latency_none_plaintext_both_p90 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_none_plaintext_both', 'p90')
    latency_v2_stats_nullvm_both_p90 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_v2-stats-nullvm_both', 'p90')
    latency_v2_stats_wasm_both_p90 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_v2-stats-wasm_both', 'p90')
    latency_v2_sd_nologging_nullvm_both_p90 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_v2-sd-nologging-nullvm_both', 'p90')
    latency_v2_sd_full_nullvm_both_p90 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_v2-sd-full-nullvm_both', 'p90')

    latency_none_mtls_base_p99 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_none_mtls_baseline', 'p99')
    latency_none_mtls_both_p99 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_none_mtls_both', 'p99')
    latency_none_plaintext_both_p99 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_none_plaintext_both', 'p99')
    latency_v2_stats_nullvm_both_p99 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_v2-stats-nullvm_both', 'p99')
    latency_v2_stats_wasm_both_p99 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_v2-stats-wasm_both', 'p99')
    latency_v2_sd_nologging_nullvm_both_p99 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_v2-sd-nologging-nullvm_both', 'p99')
    latency_v2_sd_full_nullvm_both_p99 = get_telemetry_mode_y_series(cur_href_links, cur_release_dates, '_v2-sd-full-nullvm_both', 'p99')

    context = {'current_release': current_release,
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
               }
    return render(request, "cur_regression.html", context=context)


# Create your views here.
def master_regression(request):
    _, _, _, master_href_links, _, master_release_dates = download.download_benchmark_csv(60)

    latency_none_mtls_base_p90_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_none_mtls_baseline', 'p90')
    latency_none_mtls_both_p90_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_none_mtls_both', 'p90')
    latency_none_plaintext_both_p90_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_none_plaintext_both', 'p90')
    latency_v2_stats_nullvm_both_p90_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_v2-stats-nullvm_both', 'p90')
    latency_v2_stats_wasm_both_p90_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_v2-stats-wasm_both', 'p90')
    latency_v2_sd_nologging_nullvm_both_p90_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_v2-sd-nologging-nullvm_both','p90')
    latency_v2_sd_full_nullvm_both_p90_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_v2-sd-full-nullvm_both', 'p90')

    latency_none_mtls_base_p99_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_none_mtls_baseline', 'p99')
    latency_none_mtls_both_p99_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_none_mtls_both', 'p99')
    latency_none_plaintext_both_p99_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_none_plaintext_both', 'p99')
    latency_v2_stats_nullvm_both_p99_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_v2-stats-nullvm_both', 'p99')
    latency_v2_stats_wasm_both_p99_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_v2-stats-wasm_both', 'p99')
    latency_v2_sd_nologging_nullvm_both_p99_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_v2-sd-nologging-nullvm_both', 'p99')
    latency_v2_sd_full_nullvm_both_p99_master = get_telemetry_mode_y_series(master_href_links, master_release_dates, '_v2-sd-full-nullvm_both', 'p99')

    context = {'latency_none_mtls_base_p90_master': latency_none_mtls_base_p90_master,
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
               }

    return render(request, "master_regression.html", context=context)


# Helpers
def get_latency_y_data_point(df, telemetry_mode, quantiles):
    y_series_data = []
    data = df.query('ActualQPS == 1000 and NumThreads == 16 and Labels.str.endswith(@telemetry_mode)')
    quantile_data = data.get(quantiles)
    if quantile_data is None or len(quantile_data) == 0:
        y_series_data.append('null')
    else:
        y_series_data.append(data[quantiles].head(1).values[0] / 1000)
    return y_series_data


def get_telemetry_mode_y_series(release_href_links, release_dates, telemetry_mode, quantiles):
    trending_data = [[]] * len(release_href_links)
    for i in range(len(release_href_links)):
        release_year = release_dates[i][0:4]
        release_month = release_dates[i][4:6]
        release_date = release_dates[i][6:]
        release_list = [release_year, release_month, release_date]

        try:
            href_parts = release_href_links[i].split("/")
            benchmark_test_id = href_parts[4]
            df = pd.read_csv(perf_data_path + benchmark_test_id + "_benchmark.csv")
        except Exception as e:
            print(e)
            trending_data[i] = release_list + ["null"]
        else:
            trending_data[i] = release_list + [get_latency_y_data_point(df, telemetry_mode, quantiles)]

    return trending_data


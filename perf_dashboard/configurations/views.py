from django.shortcuts import render
import pandas as pd


# Create your views here.
def configuration(request):
    # Parse data for the master
    df = pd.read_csv("/Users/carolynprh/PycharmProjects/perf_dashboard/perf_data/tmp3f9jejbf.csv")
    latency_mixer_base_p90_master = get_latency_y_series(df, '_mixer_base', 'p90')
    latency_mixer_serveronly_p90_master = get_latency_y_series(df, '_mixer_serveronly', 'p90')
    latency_mixer_both_p90_master = get_latency_y_series(df, '_mixer_both', 'p90')
    latency_nomixer_serveronly_p90_master = get_latency_y_series(df, '_nomixer_serveronly', 'p90')
    latency_nomixer_both_p90_master = get_latency_y_series(df, '_nomixer_both', 'p90')
    latency_v2_serveronly_p90_master = get_latency_y_series(df, 'nullvm_serveronly', 'p90')
    latency_v2_both_p90_master = get_latency_y_series(df, 'nullvm_both', 'p90')

    context = {'latency_mixer_base_p90_master': latency_mixer_base_p90_master,
               'latency_mixer_serveronly_p90_master': latency_mixer_serveronly_p90_master,
               'latency_mixer_both_p90_master': latency_mixer_both_p90_master,
               'latency_nomixer_serveronly_p90_master': latency_nomixer_serveronly_p90_master,
               'latency_nomixer_both_p90_master': latency_nomixer_both_p90_master,
               'latency_v2_serveronly_p90_master': latency_v2_serveronly_p90_master,
               'latency_v2_both_p90_master': latency_v2_both_p90_master,
               }
    return render(request, "configurations.html", context=context)


# Helpers
def get_latency_y_series(df, mixer_mode, quantiles):
    y_series_data = []
    for thread in [2, 4, 8, 16, 32, 64]:
        data = df.query('ActualQPS == 1000 and NumThreads == @thread and Labels.str.endswith(@mixer_mode)')
        if not data[quantiles].head().empty:
            y_series_data.append(data[quantiles].head(1).values[0])
        else:
            y_series_data.append('null')
    print(y_series_data)
    return y_series_data
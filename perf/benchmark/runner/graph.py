from bokeh.plotting import figure, output_file, show
import pandas as pd
import os
import numpy as np
from bokeh.io import output_notebook
from bokeh.models import ColumnDataSource, HoverTool
from bokeh.models.tools import CustomJSHover
from bokeh.palettes import Dark2_5 as palette
import itertools #for cycling through colors 
from bokeh.models import Legend
import sys
import argparse 

# generate_chart displays numthreads vs. metric, writes to interactive HTML
def generate_chart(csv, metric):
    print("Generating chart, metric=%s, csv=%s ..." % (metric, csv))

    # valid options = latency, memory, cpu
    valid_metrics = {"p50": "p50", "p90": "p90", "p99": "p99", "mem": "mem_MB_max_fortioserver_deployment_proxy", "cpu": "cpu_mili_max_fortioserver_deployment_proxy"}

    if metric is None:
        sys.exit('need metric')
    if metric not in valid_metrics: 
        sys.exit("invalid metric")
    if csv is None:
        sys.exit('need CSV file') 
    
    m = valid_metrics[metric] #the CSV label

    # 1. read CSV to pandas dataframe 
    df = pd.read_csv(csv, index_col=None, header=0)
    df["Labels"] = [ x.split('_', 6)[-1] for x in df['Labels']]

    # 2. generate series to plot (x=connections (numthreads), y=metric) 
    c, series = get_series(df, m) 

    # 3. generate title 
    qps=df.at[0, 'ActualQPS'] 
    seconds=df.at[0, 'ActualDuration']
    title="Istio {}, {} QPS over {}s".format(metric, qps, seconds)

    # 4. prep file-write 
    fn = "".join(title.split())
    f = "/tmp/" + fn + ".html" 
    output_file(f) 

    # 5. create chart -> save as interactive HTML 
    p = build_chart(title, metric, c, series)  
    show(p)
    print("HTML graph saved at %s" %  f)


# get_series processes metric for different test modes
def get_series(df, metric): 
    modes = {'^serveronly': 'serveronly', "nomix.*_serveronly": "nomixer_serveronly", "nomix.*_both": "nomixer_both", "base": "base", "^both": "both"}
    series = {}
    for m, k in modes.items():
        rows = df[df.Labels.str.contains(m)]
        vals = list(rows[metric]) 

        # if metric is latency, convert microseconds to milliseconds 
        if metric.startswith("p"):
            print("converting CSV microseconds to milliseconds...")
            vals = [v/1000 for v in vals]
        # reverse to match sorted numthreads, below
        vals.reverse()
        series[k] = vals 

    # get x axis (connections) 
    c = list(rows.NumThreads)
    c.sort()

    # only include modes that were in the input CSV 
    # (if nomixer not included, don't attempt to plot it)
    useries = {}
    for k, v in series.items():
        if len(v) > 0: 
            useries[k] = v 
    series = useries 

    return c, series 

# build_chart creates a bokeh.js plot from data 
def build_chart(title, metric, c, series): 
    # generate y-axis label with units 
    if metric.startswith('p'):
        label = metric = " latency, milliseconds"
    else:
        if metric == "mem":
            label = "max memory usage, server proxy (MB)"
        else:
            label = "max CPUs, server proxy (millicores)" 


    # show metric value on hover
    TOOLTIPS = [(metric, '$data_y')]

    # create plot 
    p = figure(
        tools="pan,box_zoom,reset,save",
        title=title,
        tooltips=TOOLTIPS,
        plot_width=1000, plot_height=600,
        x_axis_label='connections', y_axis_label=label  
    )

    # format axes 
    p.title.text_font_size = '22pt'
    p.xaxis.minor_tick_line_color = None  # turn off x-axis minor ticks
    p.yaxis.minor_tick_line_color = None  # turn off y-axis minor ticks
    p.xaxis.axis_label_text_font_size = "15pt"
    p.yaxis.axis_label_text_font_size = "15pt"
    p.xaxis.major_label_text_font_size = "13pt"
    p.yaxis.major_label_text_font_size = "13pt"
    p.xaxis.ticker = c  # only show a tick per numthread (not a continuous variable)

    # use a different color for each series (both, baseline, etc.)
    colors = itertools.cycle(palette) 
    for mode, val in series.items(): 
        col = next(colors)
        p.line(c, val, line_color=col)
        p.circle(c, val, legend=mode, size=10, fill_color=col)

    p.legend.location = "top_left"

    return p


def main(argv):
    args = getParser().parse_args(argv)
    return generate_chart(args.csv, args.metric) 


def getParser():
    parser = argparse.ArgumentParser("Istio Performance Graph Generator")
    parser.add_argument("csv", help="csv file", default="")
    parser.add_argument("metric", help="one of: p50, p90, p99, mem, cpu", default="")
    return parser


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))

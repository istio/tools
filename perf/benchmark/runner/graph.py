from bokeh.plotting import figure, output_file, show
import pandas as pd
import os
import numpy as np
from bokeh.io import output_notebook
# from IPython.display import display, HTML
from bokeh.models import ColumnDataSource, HoverTool
from bokeh.models.tools import CustomJSHover
from bokeh.palettes import Dark2_5 as palette
import itertools #for cycling through colors 
from bokeh.models import Legend

import sys
import argparse 

"""

"""

def generate_chart(csvs, metric):
    valid_metrics = {"p50": "p50", "p90": "p90", "p99": "p99"}#, "mem": "mem_MB_max_fortioserver_deployment_proxy", "cpu": "cpu_mili_avg_fortioserver_deployment_proxy"}

    if metric is None:
        sys.exit('need metric')
    if metric not in valid_metrics: 
        sys.exit("invalid metric")
    if csvs is None:
        sys.exit('need one or more CSV files') 
    
    m = valid_metrics[metric] #the CSV field to plot 
    files=csvs.split(",") 

    # 1. read in all rows in both files  
    df = read_csv(files) 

    # 2. generate series to plot (x=connections, y=metric) 
    c, series = get_series(df, m) 

    # 3. generate title 
    qps=df.at[0, 'ActualQPS'] 
    seconds=df.at[0, 'ActualDuration']
    title="Istio {}, {} QPS over {} seconds".format(metric, qps, seconds)

    # 4. prep file-write 
    fn = "".join(title.split())
    f = "/tmp/" + fn + ".html" 

    output_file(f) 

    # 5. create chart 
    p = build_chart(title, metric, c, series)  
    print("HTML graph saved at %s" %  f)

# returns pandas DF 
def read_csv(all_files):
    li = []
    for filename in all_files:
        df = pd.read_csv(filename, index_col=None, header=0)
        li.append(df)
    frame = pd.concat(li, axis=0, ignore_index=True)
    frame["Labels"] = [ x.split('_', 6)[-1] for x in frame['Labels']]
    return frame 

def get_series(df, metric): 
    # display(df)
    modes = {'^serveronly': 'serveronly', "nomix_serveronly": "nomix_serveronly", "nomix_both": "nomix_both", "base": "base", "^both": "both"}
    series = {}
    for m, k in modes.items():
        print(k)
        rows = df[df.Labels.str.contains(m)]
        # display(rows)
        vals = list(rows[metric]) 
        vals = [v/1000 for v in vals]
        vals.reverse()
        series[k] = vals 

    # get x axis (connections) 
    c = list(rows.NumThreads)
    c.sort()
    print(c)
    print(series)
    return c, series 

def build_chart(title, metric, c, series): 
    TOOLTIPS = [(metric, '$data_y')]
    p = figure(
        tools="pan,box_zoom,reset,save",
        title=title,
        tooltips=TOOLTIPS,
        plot_width=1000, plot_height=600,
        x_axis_label='connections', y_axis_label='P90 latency, millis'
    )

    # format axes 
    p.title.text_font_size = '22pt'
    p.xaxis.minor_tick_line_color = None  # turn off x-axis minor ticks
    p.yaxis.minor_tick_line_color = None  # turn off y-axis minor ticks
    p.xaxis.axis_label_text_font_size = "15pt"
    p.yaxis.axis_label_text_font_size = "15pt"
    p.xaxis.major_label_text_font_size = "13pt"
    p.yaxis.major_label_text_font_size = "13pt"
    p.xaxis.ticker = c

    # plot a different color for each series 
    colors = itertools.cycle(palette) 
    for mode, val in series.items(): 
        col = next(colors)
        p.line(c, val, line_color=col)
        p.circle(c, val, legend=mode, size=10, fill_color=col)

    p.legend.location = "top_left"
    return p


def main(argv):
    args = getParser().parse_args(argv)
    return generate_chart(args.metric, args.csvs) 


def getParser():
    parser = argparse.ArgumentParser("Istio Performance Graph Generator")
    parser.add_argument(
        "metric", help="Which latency metric to display: p50, p90, or p99", default="")
    parser.add_argument("csvs", help="one or more CSVs, comma-separated", default="")
    return parser


if __name__ == "__main__":
    import sys
    sys.exit(main(sys.argv[1:]))

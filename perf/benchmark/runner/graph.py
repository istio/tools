from bokeh.plotting import figure, output_file, show
import pandas as pd
import os
import numpy as np
from bokeh.io import output_notebook
from IPython.display import display, HTML


from bokeh.models import ColumnDataSource, HoverTool
from bokeh.models.tools import CustomJSHover
from bokeh.palettes import Dark2_5 as palette
import itertools #for cycling through colors 
from bokeh.models import Legend


# In[2]:


output_notebook()


# In[3]:


df = pd.read_csv("v12perf.csv")


# In[5]:


# df = df.drop(df.index[18])
# df = df.drop(df.index[18])
display(df)


# In[6]:


# format title 
percentile="p90"
qps=df.at[0, 'ActualQPS'] 
seconds=df.at[0, 'ActualDuration']

title="Istio {} Latency, {} QPS over {} seconds".format(percentile, qps, seconds)
print(title)


# In[48]:


# get series (y axis)

# test modes 
modes = {'[0-9]+_serveronly': 'serveronly', "nomix_serveronly": "nomix_serveronly", "nomix_both": "nomix_both", "base": "base", "[0-9]+_both": "both"}
series = {}
for m, k in modes.items():
    print(m)
    rows = df[df.Labels.str.contains(m)]
    display(rows)
    vals = list(rows[percentile]) # get latency values for this mode - NOTE -reversing
    vals = [v/1000 for v in vals]
    vals.reverse()
    series[k] = vals 

# get x axis (connections) 
c = list(rows.NumThreads)
c.sort()
print(c)

print(series)


# In[49]:


# output to static HTML file
fn = "".join(title.split())
output_file(fn + ".html")

TOOLTIPS = [('latency (millis)', '$data_y')]


# create a new plot
p = figure(
   tools="pan,box_zoom,reset,save",
    title=title,
    tooltips=TOOLTIPS,
    plot_width=1000, plot_height=600,
   x_axis_label='connections', y_axis_label='P90 latency, millis'
)


# turn off minor ticks; increase font sizes
p.title.text_font_size = '22pt'
p.xaxis.minor_tick_line_color = None  # turn off x-axis minor ticks
p.yaxis.minor_tick_line_color = None  # turn off y-axis minor ticks
p.xaxis.axis_label_text_font_size = "15pt"
p.yaxis.axis_label_text_font_size = "15pt"
p.xaxis.major_label_text_font_size = "13pt"
p.yaxis.major_label_text_font_size = "13pt"


# x axis should reflect numthreads 
p.xaxis.ticker = c

# plot latency data for each mode 
colors = itertools.cycle(palette) 

for mode, latency in series.items(): 
    col = next(colors)
    p.line(c, latency, line_color=col)
    p.circle(c, latency, legend=mode, size=10, fill_color=col)


# In[50]:



# move legend
p.legend.location = "top_left"


# In[51]:


show(p)


# In[23]:


df["Labels"] = [ x.split('_', 6)[-1] for x in df['Labels']]
df["p90"] = df["p90"] / 1000.0  #convert latency from ms to seconds 


# In[24]:


for label in df["Labels"].unique():
    ll = df[df["Labels"]==label]
    ll.sort_values(by=["NumThreads"])
    
    p = figure(
   tools="pan,box_zoom,reset,save",
        title="p90 Latency with increasing # connections",
   x_axis_label='connections', y_axis_label='P90 latency, seconds'
)
    
    p.circle(ll["NumThreads"], ll["p90"], name=label, legend=label)


# In[ ]:





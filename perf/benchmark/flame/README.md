# Creating cpu flame graphs for Istio / Envoy

![example](example_flame_graph/example_flagmegraph.svg)

[Flame graph](http://www.brendangregg.com/perf.html#FlameGraphs) shows how much time was spent in a particular function.
1. The width of a stack element (flame) indicates the relative time spent in the function.
1. Call stacks are plotted vertically.
1. Colors are arbitrary.
1. Function names are sorted left to right.

This document shows how to gather performance data from via the `perf` container.

## Setup the perf container

Enable `profilingMode` in [values.yaml](../values.yaml). This will end up adding the perf
container to the server and client pods, which both will be running on separate nodes.

Flame graphs are created from data collected using linux `perf_events` by the `perf` and [BCC tools](https://github.com/iovisor/bcc).

## Obtaining flame graphs

Flame graphs can be produced via `runner.py`, and will be stored in `flame/flameoutput`.

A few sample command lines. `{duration}` will be replaced by
whatever was passed for `--duration` to runner.py. `{sidecar_pid}` will
be replaced by `runner.py` with the process id of the Envoy sidecar.

It is valid to omit `{sidecar_pid}` in `--custom_profiling_command`.
This may be useful for machine-wide profiling or arbitrary processes.

```bash
runner/runner.py --conn 20 --qps 10000 --duration 100 --custom_profiling_command="profile-bpfcc -df {duration} -p {sidecar_pid}" --custom_profiling_name="bcc-oncputime-sidecar"

runner/runner.py --conn 20 --qps 10000 --duration 100 --serversidecar --custom_profiling_command="offcputime-bpfcc -df {duration} -p {sidecar_pid}" --custom_profiling_name="bcc-offcputime-sidecar"

runner/runner.py --conn 20 --qps 10000 --duration 100 --serversidecar --custom_profiling_command="offwaketime-bpfcc -df {duration} -p {sidecar_pid}" --custom_profiling_name="bcc-offwaketime-sidecar"

runner/runner.py --conn 20 --qps 10000 --duration 100 --serversidecar --custom_profiling_command="wakeuptime-bpfcc -f -p {sidecar_pid} {duration}" --custom_profiling_name="bcc-wakeuptime-sidecar"

runner/runner.py --conn 20 --qps 10000 --duration 100 --serversidecar --custom_profiling_command="perf record -F 99 -g -p {sidecar_pid} -- sleep {duration} && perf script | ~/FlameGraph/stackcollapse-perf.pl | c++filt -n" --custom_profiling_name="perf-oncputime-sidecar"
```

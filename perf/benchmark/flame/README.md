Creating cpu flame graphs for Istio / Envoy
---------------------------------------

![example](example_flagmegraph.svg)

[Flame graph](http://www.brendangregg.com/perf.html#FlameGraphs) shows how much time was spent in a particular function.
1. The width of a stack element (flame) indicates the relative time spent in the function.
2. Call stacks are plotted vertically.
3. Colors are arbitrary.
4. Function names are sorted left to right.


This document shows how to gather performance data from within the `istio-proxy` container.

Setup Perf tool
---------------
Flame graphs are created from data collected using linux `perf_events` by the `perf` tool.

1. Ensure that `perf` is installed within the container.
   Since `istio-proxy` container does not allow installation of new packages, build a new docker image.

    ```
    FROM gcr.io/istio-release/proxyv2:release-1.0-20180810-09-15
    # Install fpm tool
    RUN  sudo apt-get update && \
        sudo apt-get -qqy install linux-tools-generic

    ```
    Build image and push docker image and use it in your deployment by adding the following annotation.
    ```
        "sidecar.istio.io/proxyImag" : <name of your image>
    ```
    This step will go away once the default debug image contains `perf` and related tools.

2. Ensure that you can run `perf record` 

    Running `perf record` from container requires the host to permit this activity. This is done by running the following command on the vm host.
    ```
    sudo sysctl kernel.perf_event_paranoid=-1
    sudo sysctl kernel.kptr_restrict=0
    ```
    This setting is very permissive so it must be used with care.

3. Copy [`get_perfdata.sh`](get_perfdata.sh) to the container and run it as follows. The following command collects samples at `177Hz` for `20s`.
    ```
    istio-proxy@fortioserver-deployment-84fcdcbcf9-47f75:/etc/istio/proxy$ ./get_perfdata.sh perf.data 20 177
    ...
    [ perf record: Woken up 1 times to write data ]
    [ perf record: Captured and wrote 0.040 MB /etc/istio/proxy/perf.data (157 samples) ]

    Wrote /etc/istio/proxy/perf.data.perf
    ...
    ```
4. Copy file out of the container
    ```
    kubectl cp <pod>:/etc/istio/proxy/perf.data.perf ./perf.data.perf -c istio-proxy
    ``` 

5. Run [`flame.sh`](flame.sh) on the downloaded file
    ```
    ./flame.sh ./perf.data.perf
    Wrote perf.svg

    Copying file://perf.svg [Content-Type=image/svg+xml]...
    / [1 files][ 89.4 KiB/ 89.4 KiB]
    Operation completed over 1 objects/89.4 KiB.
    ```

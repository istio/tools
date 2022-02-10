# Control Plane Testing

This directory tests the stability of istiod when security policies are deployed in bulk. The intent of this test is to analyse the cpu load and memory usage of istiod. The polices are being deployed in incremental manner where we initially deploy just 3 Authorization and Request Authentication Policies and later we deploy 100 of these policies and finally we deploy 1000 Authorization policies. Prometheus and Grafana is integrated to analyze the ups and down in cpu load and memory usage.

## Setup

1) First you need to setup the alertmanger, prometheus, grafana on your cluster which you can do by following this link - `https://github.com/istio/tools/tree/master/perf/stability#readme`

2) Run the following commands in different terminals.

    `istioctl dashboard prometheus -n istio-promethues`

    `istioctl dashboard grafana -n istio-system`

3) Grafana dashboard is running on `localhost:3000`, you can check the istio-controlplane dashboard there to see the various metrics.

4) Run the following command to run the control plane tests.

    `./test.sh`



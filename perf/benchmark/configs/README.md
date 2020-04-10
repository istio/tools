# Add new config to benchmark pipeline

Currently we are running benchmark test towards different configs as [prow job](https://prow.istio.io/job-history/istio-prow/logs/daily-performance-benchmark)

To add a new config to this pipeline, we need to add a new directory under [configs folder](https://github.com/istio/tools/tree/master/perf/benchmark/configs/istio), where we can define config parameters structured as below:

- installation.yaml: install Istio with this IstioOperator overlay file on top of istioctl built-in default profile and [perf testing default overlay](https://github.com/istio/tools/tree/master/perf/istio-install/istioctl_profiles/default.yaml)
- cpu_mem.yaml: if provided, run cpu, memory test with this config
- latency.yaml: if provided, run latency test with this config
- prerun.sh: prerun hook we want to run before test
- postrun.sh: postrun hook we want to run after test

An example of recently added new config: [plaintext](https://github.com/istio/tools/tree/master/perf/benchmark/configs/istio/plaintext)

Note: we have a [run_perf_test.conf](./istio/run_perf_test.conf) file to control the disable and enable of the set of perf test configs.


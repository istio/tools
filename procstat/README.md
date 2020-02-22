# Proc stat sampler

## Run tests

The following command runs a couple of unit tests for the proc stat sampling system.

```bash
python3 -m unittest discover -v tests/
```

## Using the sampler

```
./main.py --help
usage: main.py [-h] [--sample-frequency SAMPLE_FREQUENCY]
               [--track-proc-name [TRACK_PROC_NAME [TRACK_PROC_NAME ...]]]
               [--dump-path DUMP_PATH]

Proc stat sampler CLI

optional arguments:
  -h, --help            show this help message and exit
  --sample-frequency SAMPLE_FREQUENCY
                        Number of samples to obtain per second. Defaults to 1
                        per second.
  --track-proc-name [TRACK_PROC_NAME [TRACK_PROC_NAME ...]]
                        Process name(s) to track, if any. Multiple allowed.
  --dump-path DUMP_PATH
                        Path where the result will be written.
```

# Transform a dump from the sampler to yaml

```
oschaaf@burst:~/code/istio/tools/procstat$ ./dump-to-yaml.py --help
usage: dump-to-yaml.py [-h] [--dump-path DUMP_PATH]

Transforms dumps from the sampler to yaml

optional arguments:
  -h, --help            show this help message and exit
  --dump-path DUMP_PATH
                        Path where the target dump resides.
```

### Sample output:

```
- cpu_percent: 2.4
  cpu_times:
    guest: 0.0
    guest_nice: 0.0
    idle: 8788185.5
    iowait: 2188.63
    irq: 0.0
    nice: 19.19
    softirq: 14.13
    steal: 0.0
    system: 765.38
    user: 5233.24
  processes: []
  timestamp: 1581979800.9612823
- cpu_percent: 0.0
  cpu_times:
    guest: 0.0
    guest_nice: 0.0
    idle: 8788225.51
    iowait: 2188.63
    irq: 0.0
    nice: 19.19
    softirq: 14.13
    steal: 0.0
    system: 765.38
    user: 5233.25
  processes: []
  timestamp: 1581979801.9625692
- cpu_percent: 0.0
  cpu_times:
    guest: 0.0
    guest_nice: 0.0
    idle: 8788265.53
    iowait: 2188.63
    irq: 0.0
    nice: 19.19
    softirq: 14.13
    steal: 0.0
    system: 765.38
    user: 5233.25
  processes: []
  timestamp: 1581979802.963791
```

## Expose the proc stat sampler output for prometheus scraping

```bash
# run in a separate terminal
./prom.py  --track nginx envoy --http-port 8000
```

# Querying the statistics

Note: output is from an early version, which only tracked internal metrics from the prometheus client lib).

```bash
curl --silent 127.00.1:8000 | head

oschaaf@burst:~/code/istio/tools/procstat$ curl --silent 127.0.0.1:8000 | head
# HELP python_gc_objects_collected_total Objects collected during gc
# TYPE python_gc_objects_collected_total counter
python_gc_objects_collected_total{generation="0"} 123.0
python_gc_objects_collected_total{generation="1"} 255.0
python_gc_objects_collected_total{generation="2"} 0.0
# HELP python_gc_objects_uncollectable_total Uncollectable object found during GC
# TYPE python_gc_objects_uncollectable_total counter
python_gc_objects_uncollectable_total{generation="0"} 0.0
python_gc_objects_uncollectable_total{generation="1"} 0.0
python_gc_objects_uncollectable_total{generation="2"} 0.0

```

## Exposing prometheus metrics in side car proxy containers.

The following script will build a standalone binary, deploy it to the benchmark
side car proxy containers, and fire up the service.

```bash
NAMESPACE=twopods-istio ./install_to_container.sh
```

## Testing if the service is running in containers

The service will listen on port 8000 by default. Hence querying that port with curl ought to output a bunch of counters in prometheus format.

``` bash
kubectl --namespace twopods-istio exec fortioclient-6b58bf5799-hkq8l -c istio-proxy curl 127.0.0.1:8000

...
cpu_times_system 6217.48
...
```


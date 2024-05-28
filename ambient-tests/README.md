# Performance Tests for Ztunnel

Performance tests for Istio Ambient, Istio Ambient + Waypoint, Istio Sidecar.
Uses [netperf](https://github.com/HewlettPackard/netperf) and [fortio](https://github.com/fortio/fortio).

## Setup

To run, first set up a cluster with two user nodes: one with the `role=server` label and one with the `role=client` label.
You can recreate this in a kind cluster using `yaml/cluster.yaml`.
If you use a kind cluster, will either have to connect it to a local registry, or upload the images using `make push-local` in the `docker` directory.
You will also have to modify the `imagePullPolicy` fields of the pods in `yaml/deploy.yaml` to `Never`.

Note that if you are using AKS to set up your cluster, make sure that you use Azure CNI as your network plugin and **DO NOT** use a network policy.
Also, make sure to attach to container registry to your cluster/that your container registry is accessible from your cluster.
This ensures that client and server pods get deployed in different servers.

Next, go into `docker/Makefile` and change the value of `CR` to your container registry and the image names in `yaml/deploy.yaml` accordingly.

You will also need a Python 3 with `matplotlib`, `pandas`, and `python-dotenv`.
Also, make sure that `python -V` is some version of Python 3.
An easy way to get this on Ubuntu is running

```bash
pipenv shell
sudo apt install python-is-python3
sudo apt install python3-pip
pip install matplotlib pandas python-dotenv
```

Install Istio Ambient.
I don't automate this process because there are many installation methods and you might be testing a custom build.
To install the latest release of Istio Ambient [install `istioctl`](https://istio.io/latest/docs/setup/getting-started/#download) and run

```bash
istioctl install --set profile=ambient
```

## Building

To build the necessary containers, inside `docker/`, run

```bash
cd docker
CR=docker.io/<YourRepoName> make build
CR=docker.io/<YourRepoName> make push-cr
# or `push-local` if running in a kind cluster.
```

if you got an error like the below,
![alt text](push_error.png)
you can tag the image locally and push to the image repository;

```bash
docker tag stjinxuan.azurecr.io/ambient-performance:latest docker.io/<YourRepoName>/ambient-performance:latest
docker push docker.io/<YourRepoName>/ambient-performance:latest
```

then update the client and server containers' images in `yaml/deploy.yaml` with your new image. then deploy the pods;

```bash
cd ~/tools/ambient-tests
./scripts/config.sh
./scripts/setup.sh
```

## Running Benchmarks

Now, update `scripts/config.sh` as desired or keep the default values, and from the `~/tools/ambient-tests` directory, run the test

```bash
./scripts/fortio/run.sh
```

then generate the csv file data;

```bash
FORTIO_RESULTS=./result/fortio scripts/fortio/gen-csv.sh
```

and finaly get the graph;

```bash
python ./scripts/fortio/graphs.py
```

This will create graphs in the `graphs/fortio` directory and put intermediate files in `results/fortio` by default.

For more configuration options, see `scripts/run.sh`.

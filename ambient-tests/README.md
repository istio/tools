# Performance Tests for Ztunnel

## Setup

To run, first set up a cluster with six user nodes: three with the `role=server` label and three with the `role=client` label.
Note that if you are using AKS to set up your cluster, make sure that you use Azure CNI as your network plugin and **DO NOT** use a network policy.
Also, make sure to attach to container registry to your cluster/that your container registry is accessible from your cluster.
This ensures that `netperf` and `netserver` pods get deployed in different servers.

Next, go into `netperf/Makefile` and change the value of `CR` to your container registry and the image names in `yaml/deploy.yaml` accordingly.

You will also need a Python 3 with `matplotlib`, `pandas`, and `python-dotenv`.
Also, make sure that `python -V` is some version of Python 3. An easy way to get this on Ubuntu is running

```bash
sudo apt install python-is-python3
sudo apt install python3-pip
pip install matplotlib pandas python-dotenv
```

Install Istio Ambient.
I don't automate this process because there are many installation methods and you might be testing a custom build.
To install the latest release of Istio Ambient [install `istioctl`](https://istio.io/latest/docs/setup/getting-started/#download) and run

```bash
istiocl install --set profile=ambient
```

## Building

To build the necessary containers, inside `netperf/`, run

```bash
make build
make push-cr
```

## Running Benchmarks

Now, update `scripts/config.sh` as desired or keep the default values, and run `./scripts/setup.sh` from the root of the repository to deploy the pods.
Once the deployments are complete, run the tests with `make run`.
This will create graphs in the `graphs/` directory and put intermediate files in `results/` by default.

For more configuration options, see `scripts/run.sh`.


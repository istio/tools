# shellcheck shell=bash

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# server-client config
export YAML_PATH=./yaml/deploy.yaml

# name of the deployments
export BENCHMARK_SERVER=server
export BENCHMARK_CLIENT=client

# where to store primary and intermediate results
export NETPERF_RESULTS=results/netperf
export FORTIO_RESULTS=results/fortio

# name of the namespaces for each mesh setup
export NS_NO_MESH=no-mesh
export NS_SIDECAR=sidecar
export NS_AMBIENT=ambient
export NS_WAYPOINT=waypoint  # ambient w/ waypoint proxy
# name of service account for server. Used for waypoint configuration.
# This is also hardcoded in deployment configuration
export SA_SERVER=server-sa

# Separator for tests runs. Doesn't really matter. Just set to something weird
export TEST_RUN_SEPARATOR="~~~~~~~~~~~~~~~~"
# How many runs of each test
export N_RUNS=2

# Extra arguments for TCP_RR and TCP_CRR tests.
# These are necessary because by default *RR tests send only one byte.
# However, Envoy proxies won't create a connection until more bytes are sent.
# This also means that reverse tests _DO NOT WORK_.
export NETPERF_RR_ARGS="-r 100"

# -P Have the data connection listen on port 35000 on the server.
# -k Output all fields in key=value from.
#    We will pick and choose later.
export NETPERF_TEST_ARGS="-P ,35000 -k all"

# -P toggles the tests banner.
#    This is very confusing because it has nothing to do with ports.
# -j to measure latency
export NETPERF_GLOBAL_ARGS="-P 0 -j"

# Args for serial tests
# -qps          As many queries per second as possible
# -c            Number of sending threads
# -payload-file POST request with Makefile as data.
#
# This configuration seems to trigger the Nagle's + delayed ack interaction.
export FORTIO_SERIAL_HTTP_ARGS="-qps -1 -c 1 -payload-size 300"

# Arguments for prallel http tests
# 3 sending threads should be enough(?)
export FORTIO_PARALLEL_HTTP_ARGS="-qps -1 -c 3"

# where to output the graphs
export NETPERF_GRAPHS="graphs/netperf"
export FORTIO_GRAPHS="graphs/fortio"

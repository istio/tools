# server-client config
export YAML_PATH=./yaml/deploy.yaml

# name of the deployments
export BENCHMARK_SERVER=server
export BENCHMARK_CLIENT=client

# where to store primary and intermediate results
export RESULTS=results

# name of the namespaces for each mesh setup
export NS_NO_MESH=no-mesh
export NS_SIDECAR=sidecar
export NS_AMBIENT=ambient

# Separator for tests runs. Doesn't really matter. Just set to something weird
export TEST_RUN_SEPARATOR="~~~~~~~~~~~~~~~~"
# How many runs of each test
export N_RUNS=2

# Extra arguments for TCP_RR and TCP_CRR tests.
# These are necessary because by default *RR tests send only one byte.
# However, Envoy proxies won't create a connection until more bytes are sent.
# This also means that reverse tests _DO NOT WORK_.
export RR_ARGS="-r 100"

# -P Have the data connection listen on port 35000 on the server.
# -k Output all fields in key=value from.
#    We will pick and choose later.
export TEST_ARGS="-P ,35000 -k all"

# -P toggles the tests banner.
#    This is very confusing because it has nothing to do with ports.
# -j to measure latency
export GLOBAL_ARGS="-P 0 -j"

# where to output the graphs
export GRAPHS="graphs"

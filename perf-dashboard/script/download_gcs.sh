#!/bin/bash

# TODO:
#BENCHMARK_PATH_NAME="benchmark_data.20191021-08.98f8f41ff524b648414089b5a7f9731479660363"
#BUCKET_NAME="istio-build/perf/${BENCHMARK_PATH_NAME}"
#DOWNLOAD_FILE_NAME="benchmark.csv"
#DESTINATION_PATH=""

WD=$(dirname "$0")
WD=$(cd "$WD"; pwd)
ROOT=$(dirname "$WD")

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print commands
set -x

gsutil cp gs://istio-build/perf/benchmark_data.20191021-08.98f8f41ff524b648414089b5a7f9731479660363/benchmark.csv "${ROOT}"/data/benchmark.csv
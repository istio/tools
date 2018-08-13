#!/bin/bash
set -e

WD=$(dirname "${0}")
WD=$(cd "${WD}" && pwd)

FILENAME=${1:?"perffilename"}
DURATION=${2:?"duration"}
FREQ=${3:-"99"}

PID=$(pgrep envoy)

# This is specific to the kernel version
# provided by `linux-tools-generic`
PERF="/usr/lib/linux-tools/4.4.0-131-generic/perf"


"${PERF}" record -o "${WD}/${FILENAME}" -F "${FREQ}" -p "${PID}" -g -- sleep "${DURATION}"

"${PERF}" script -i "${WD}/${FILENAME}" --demangle > "${WD}/${FILENAME}.perf"

echo "Wrote ${WD}/${FILENAME}.perf"

#!/bin/bash
set -e

WD=$(dirname "${0}")
WD=$(cd "${WD}" && pwd)

FLAMEDIR="${WD}/FlameGraph"

if ! which c++filt > /dev/null; then
    echo "Install c++filt to demangle symbols"
    exit 1
fi 

cd "${WD}" || exit -1

if [[ ! -d ${FLAMEDIR} ]];then
    git clone https://github.com/brendangregg/FlameGraph
fi

# Given output of `perf script` produce a flamegraph
FILE=${1:?"perf script output"}
FILENAME=$(basename "${FILE}")
BASE=$(echo "${FILENAME}" | cut -d '.' -f 1)
SVGNAME="${BASE}.svg"

"${FLAMEDIR}/stackcollapse-perf.pl" "${FILE}" | c++filt -n | "${FLAMEDIR}/flamegraph.pl" --cp > "${SVGNAME}"

echo "Wrote ${SVGNAME}"
if [[ ! -z "${BUCKET}" ]];then
    gsutil cp "${SVGNAME}" "${BUCKET}"
fi

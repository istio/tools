BASE := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GOPATH = $(shell cd ${BASE}/../../..; pwd)
TOP ?= $(GOPATH)

${GOPATH}/src/istio.io/istio:
	mkdir -p $GOPATH/src/istio.io
	git clone https://github.com/istio/istio.git ${GOPATH}/src/istio.io/istio

${GOPATH}/src/istio.io/installer:
	mkdir -p $GOPATH/src/istio.io/installer
	git clone https://github.com/istio/installer.git ${GOPATH}/src/istio.io/installer

init: ${GOPATH}/src/istio.io/istio ${GOPATH}/src/istio.io/installer
	mkdir -p ${GOPATH}/src/istio.io/istio

test: init
	$(MAKE) -C ${GOPATH}/src/istio.io/installer test $(MAKEFLAGS)

check-stability:
	./metrics/check_metrics.py

lint:
	@scripts/check_license.sh
	@scripts/run_golangci.sh
	@scripts/check_dockerfiles.sh

fmt:
	@scripts/run_gofmt.sh

fmtpy-checkandupdate:
	@scripts/check_pyfmt.sh true

fmtpy-checkonly:
	@scripts/check_pyfmt.sh false

include Makefile.common.mk
include perf/stability/stability.mk

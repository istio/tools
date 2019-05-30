BASE := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GOPATH = $(shell cd ${BASE}/../../..; pwd)
TOP ?= $(GOPATH)

${GOPATH}/src/istio.io/istio:
	mkdir -p $GOPATH/src/istio.io
	git clone https://github.com/istio/istio.git ${GOPATH}/src/istio.io/istio

${GOPATH}/src/github.com/istio-ecosystem/istio-installer:
	mkdir -p $GOPATH/src/github.com/istio-ecosystem
	git clone https://github.com/istio-ecosystem/istio-installer.git ${GOPATH}/src/github.com/istio-ecosystem/istio-installer

init: ${GOPATH}/src/istio.io/istio ${GOPATH}/src/github.com/istio-ecosystem/istio-installer
	mkdir -p ${GOPATH}/src/istio.io/istio

test: init
	$(MAKE) -C ${GOPATH}/src/github.com/istio-ecosystem/istio-installer test $(MAKEFLAGS)

check-stability:
	./metrics/check_metrics.py

lint:
	@scripts/check_license.sh
	@scripts/run_golangci.sh

fmt:
	@scripts/run_gofmt.sh

include Makefile.common.mk
include perf/stability/stability.mk

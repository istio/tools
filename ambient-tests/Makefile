.PHONY: docker-build run-netperf

run-netperf:
	./scripts/netperf/run.sh
	./scripts/netperf/gen-csv.sh
	python ./scripts/netperf/graphs.py

run-fortio:
	./scripts/fortio/run.sh
	./scripts/fortio/gen-csv.sh
	python ./scripts/fortio/graphs.py

docker-build:
	make build push -C netperf


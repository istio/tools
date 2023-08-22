.PHONY: docker-build run

run:
	./scripts/run.sh
	./scripts/gen_csv.sh
	python ./scripts/graphs.py

docker-build:
	make build push -C netperf


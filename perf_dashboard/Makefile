VERSION := $(shell date +v%Y%m%d)-$(shell git rev-parse --short HEAD)

HUB ?= gcr.io/istio-testing

CONTAINER_CLI ?= docker
IMG_NAME ?= perf-dashboard
IMG_LATEST := $(HUB)/$(IMG_NAME):latest
IMG_VERSION := $(HUB)/$(IMG_NAME):$(VERSION)

PROJECT ?= istio-testing
CLUSTER ?= perf-dashboard
ZONE    ?= us-central1-c

export KUBECONFIG := $(shell mktemp)

get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

build:
	@$(CONTAINER_CLI) build --no-cache --tag "$(IMG_LATEST)" --tag "$(IMG_VERSION)" .

run:
	@$(CONTAINER_CLI) run -p 8000:8000 --rm --name "$(IMG_NAME)" "$(IMG_LATEST)"

push:
	@$(CONTAINER_CLI) push "$(IMG_LATEST)"
	@$(CONTAINER_CLI) push "$(IMG_VERSION)"

deploy_only: get-cluster-credentials
	@helm template "perf-dashboard" deploy/perf-dashboard --set version="$(VERSION)" | kubectl apply -f -
	@echo "Deployed perf-dashboard:$(VERSION) to project:$(PROJECT) cluster:$(CLUSTER)"

deploy: build push deploy_only

.PHONY: get-cluster-credentials build push deploy_only deploy

VERSION := $(shell date +v%Y%m%d)-$(shell git rev-parse --short HEAD)

HUB ?= gcr.io/istio-testing/perf-dashboard

IMG_LATEST := $(HUB):latest
IMG_VERSION := $(HUB):$(VERSION)

PROJECT ?= istio-testing
CLUSTER ?= perf-dashboard
ZONE    ?= us-central1-c

export KUBECONFIG := $(shell mktemp)

get-cluster-credentials:
	gcloud container clusters get-credentials "$(CLUSTER)" --project="$(PROJECT)" --zone="$(ZONE)"

build:
	@docker build --no-cache --tag "$(IMG_LATEST)" --tag "$(IMG_VERSION)" .

push:
	@docker push "$(IMG_LATEST)"
	@docker push "$(IMG_VERSION)"

deploy_only: get-cluster-credentials
	@helm template "perf-dashboard" deploy/perf-dashboard --set version="$(VERSION)" | kubectl apply -f -
	@echo "Deployed perf-dashboard:$(VERSION) to project:$(PROJECT) cluster:$(CLUSTER)"

deploy: build push deploy_only

.PHONY: get-cluster-credentials build push deploy_only deploy
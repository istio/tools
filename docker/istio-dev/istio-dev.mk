# Copyright 2019 Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DEV_IMAGE_NAME = istio/dev:$(USER)
DEV_CONTAINER_NAME = istio-dev

# Build a dev environment Docker image.
docker/istio-dev/image-built: docker/istio-dev/Dockerfile
	@echo "building \"$(DEV_IMAGE_NAME)\" Docker image"
	@docker build \
		--build-arg user="${shell id -un}" \
		--build-arg group="${shell id -gn}" \
		--build-arg uid="${shell id -u}" \
		--build-arg gid="${shell id -g}" \
		--tag "$(DEV_IMAGE_NAME)" - < docker/istio-dev/Dockerfile
	@touch $@

# Start a dev environment Docker container.
.PHONY = dev-shell clean-dev-shell
dev-shell: docker/istio-dev/image-built
	@if test -z "$(shell docker ps -a -q -f name=$(DEV_CONTAINER_NAME))"; then \
	    echo "starting \"$(DEV_CONTAINER_NAME)\" Docker container"; \
		docker run --detach \
			--name "$(DEV_CONTAINER_NAME)" \
			--volume "$(GOPATH)/out:/home/$(USER)/go/out:consistent" \
			--volume "$(GOPATH)/pkg:/home/$(USER)/go/pkg:consistent" \
			--volume "$(GOPATH)/src:/home/$(USER)/go/src:consistent" \
			--volume "$(HOME)/.config/gcloud:/home/$(USER)/.config/gcloud:cached" \
			--volume "$(HOME)/.kube:/home/$(USER)/.kube:cached" \
			--volume /var/run/docker.sock:/var/run/docker.sock \
			"$(DEV_IMAGE_NAME)" \
			'while true; do sleep 60; done';  fi
	@echo "executing shell in \"$(DEV_CONTAINER_NAME)\" Docker container"
	@docker exec --tty --interactive "$(DEV_CONTAINER_NAME)" /bin/bash

clean-dev-shell:
	docker rm -f "$(DEV_CONTAINER_NAME)" || true
	if test -n "$(shell docker images -q $(DEV_IMAGE_NAME))"; then \
		docker rmi -f "$(shell docker images -q $(DEV_IMAGE_NAME))" || true; fi
	rm -f docker/istio-dev/image-built

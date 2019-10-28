# Copyright Istio Authors
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

build:
	@go build ./...

test:
	@go test -race ./...

check-stability:
	./metrics/check_metrics.py

MARKDOWN_LINT_WHITELIST=mysite.com/mypage.html,github.com/istio/istio/releases/download/untagged-c41cff3404b8cc79a97e/istio-1.1.0-rc.0-linux.tar.gz,localhost

lint: lint-all

fmt: format-go tidy-go format-python

gen: tidy-go mirror-licenses

gen-check: gen check-clean-repo

containers:
	@gcloud auth configure-docker -q	# enable docker to authenticate with gcr.io, needed for prow
	@cd docker/build-tools && ./build-and-push.sh

include common/Makefile.common.mk
include perf/stability/stability.mk

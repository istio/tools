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

"""Common constants used throughout the runner module."""

import datetime

DEFAULT_NAMESPACE = 'default'
MONITORING_NAMESPACE = 'monitoring'
ISTIO_NAMESPACE = 'istio-system'
SERVICE_GRAPH_NAMESPACE = 'service-graph'

DEFAULT_NODE_POOL_NAME = 'default-pool'
SERVICE_GRAPH_NODE_POOL_NAME = 'service-graph-pool'
CLIENT_NODE_POOL_NAME = 'client-pool'
CLIENT_NAME = 'client'
CLIENT_PORT = 8080
SERVICE_GRAPH_SERVICE_SELECTOR = 'role=service'
SERVICE_PORT = 8080
ISTIO_INGRESS_GATEWAY_PORT = 80

PROMETHEUS_SCRAPE_INTERVAL = datetime.timedelta(seconds=30)

ISTIO_TELEMETRY_PORT = 42422

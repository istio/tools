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

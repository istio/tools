from prometheus import Query, Alarm

graceful_shutdown = [
    Query(
        'Graceful Shutdown: 5xx Requests/s',
        'sum(rate(istio_requests_total{destination_service="httpbin.graceful-shutdown.svc.cluster.local", source_app="client", response_code=~"5.."}[10m]))',
        Alarm(
            lambda error_rate: error_rate > 0,
            'There were 5xx errors. Requests may be getting dropped.'
        )
    ),
    Query(
        'Graceful Shutdown: Total Requests/s',
        'sum(rate(istio_requests_total{destination_service="httpbin.graceful-shutdown.svc.cluster.local", source_app="client"}[10m]))',
        Alarm(
            lambda qps: qps < 18,
            'Not enough requests sent; expect at least 18. Service may be having issues.'
        )
    ),
]

external_traffic = [
    Query(
        'External Traffic: Total requests',
        'sum(rate(istio_requests_total{destination_service="fortio-server.allow-external-traffic-b.svc.cluster.local"}[10m]))',
        Alarm(
            lambda qps: qps < 250,
            'Not enough requests sent; expect at least 250. Service may be having issues.'
        )
    )
    # Cross namespace metrics are not recorded
]

queries = graceful_shutdown + external_traffic

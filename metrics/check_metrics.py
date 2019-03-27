#!/usr/bin/env python3
from prometheus import Query, Alarm, Prometheus
import subprocess
import unittest
import os

__unittest = True  # Will hide traceback, making test output cleaner


def assert_empty(l, msg):
    assert len(l) == 0, msg


class TestAlarms(unittest.TestCase):
    def test_graceful_shutdown(self):
        queries = [
            Query(
                'Graceful Shutdown: 5xx Requests/s',
                'sum(rate(istio_requests_total{destination_service="httpbin.graceful-shutdown.svc.cluster.local", source_app="client", response_code=~"5.."}[10m]))',
                Alarm(
                    lambda error_rate: error_rate > 0,
                    'There were 5xx errors. Requests may be getting dropped.'
                ),
                None
            ),
            Query(
                'Graceful Shutdown: Total Requests/s',
                'sum(rate(istio_requests_total{destination_service="httpbin.graceful-shutdown.svc.cluster.local", source_app="client"}[10m]))',
                Alarm(
                    lambda qps: qps < 18,
                    'Not enough requests sent; expect at least 18. Service may be having issues.'
                ),
                None
            ),
        ]
        self.run_queries(queries)

    def test_external_traffic(self):
        queries = [
            Query(
                'External Traffic: Total requests',
                'sum(rate(istio_requests_total{destination_service="fortio-server.allow-external-traffic-b.svc.cluster.local"}[10m]))',
                Alarm(
                    lambda qps: qps < 250,
                    'Not enough requests sent; expect at least 250. Service may be having issues.'
                ),
                None
            )
            # Cross namespace metrics are not recorded
        ]
        self.run_queries(queries)

    def test_redis(self):
        queries = [
            Query(
                'Redis: error rate',
                'sum(rate(stability_outgoing_requests_total{source="redis-client", succeeded="False"}[5m]))/sum(rate(stability_outgoing_requests_total{source="redis-client"}[5m]))',
                Alarm(
                    lambda errs: errs > 0,
                    'Error rate too high, expected no errors'
                ),
                'sum(stability_test_instances{test="redis"})'
            )
        ]
        self.run_queries(queries)

    def setUp(self):
        port = os.environ.get("PROM_PORT", "9990")
        self.port_forward = subprocess.Popen([
            'kubectl',
            '-n', 'istio-system',
            'port-forward',
            'deployment/prometheus',
            '%s:9090' % port
        ], stdout=subprocess.PIPE)

        self.port_forward.stdout.readline()  # Wait for port forward to be ready

        self.prom = Prometheus('http://localhost:%s/' % port)

    def tearDown(self):
        self.port_forward.stdout.close()  # Wait for port forward to be ready
        self.port_forward.terminate()
        self.port_forward.wait()

    def run_queries(self, queries):
        for query in queries:
            with self.subTest(name=query.description):
                if query.running_query:
                    if self.prom.fetch_value(query.running_query) == 0:
                        self.skipTest("Test is not running")
                errors = self.prom.run_query(query)
                message = 'Alarms Triggered:'
                for e in errors:
                    message += '\n- ' + e
                assert_empty(errors, message)


if __name__ == '__main__':
    unittest.main(verbosity=2)

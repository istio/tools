#!/usr/bin/env python3
from prometheus import Query, Alarm, Prometheus
import subprocess
import unittest
import os

__unittest = True  # Will hide traceback, making test output cleaner

CPU_MILLI = 1000
MEM_MB = 1 / (1024 * 1024)


def assert_empty(l, msg):
    assert len(l) == 0, msg


def find_prometheus():
    """Prometheus can be in different locations depending on if we use operator
    or default Istio install."""
    try:
        subprocess.check_output(
            ['kubectl', 'get', '-n', 'istio-prometheus',
                'deployment/istio-operator'],
            stderr=subprocess.DEVNULL
        )
        return "istio-prometheus", "statefulset/prometheus-istio-prometheus"
    except subprocess.CalledProcessError:
        return "istio-system", "deployment/prometheus"

# count(count_values("value", envoy_cluster_manager_cds_version))
# this values changes, watch out this change.
def standard_queries(namespace, cpu_lim=50, mem_lim=64):
    """Standard queries that should be run against all tests."""
    return [
        Query(
            '%s: 5xx Requests/s' % namespace,
            'sum(rate(istio_requests_total{reporter="destination", destination_service_namespace=~"%s", response_code=~"5.."}[1m]))' % namespace,
            Alarm(
                lambda error_rate: error_rate > 0,
                'There were 5xx errors.'
            ),
            None
        ),
    ]


class TestAlarms(unittest.TestCase):
    def test_pilot(self):
        queries = [
            Query(
                "Pilot: XDS rejections",
                'pilot_total_xds_rejects',
                Alarm(
                    lambda errors: errors > 0,
                    'There should not be any rejected XDS pushes'
                ),
                None
            )
        ]
        self.run_queries(queries)

    @classmethod
    def setUpClass(self):
        port = os.environ.get("PROM_PORT", "9990")
        namespace, deployment = find_prometheus()
        self.port_forward = subprocess.Popen([
            'kubectl',
            '-n', namespace,
            'port-forward',
            deployment,
            '%s:9090' % port
        ], stdout=subprocess.PIPE)

        self.port_forward.stdout.readline()  # Wait for port forward to be ready

        self.prom = Prometheus('http://localhost:%s/' % port)

    @classmethod
    def tearDownClass(self):
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

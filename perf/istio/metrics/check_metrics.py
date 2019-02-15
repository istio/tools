#!/usr/bin/env python3
from prometheus import Query, Alarm, Prometheus
import subprocess
import unittest
import test_cases
import os

__unittest = True  # Will hide traceback, making test output cleaner


def assert_empty(l, msg):
    assert len(l) == 0, msg


class TestAlarms(unittest.TestCase):
    port = os.environ.get("PROM_PORT", "9990")

    def setUp(self):
        self.port_forward = subprocess.Popen([
            'kubectl',
            '-n', 'istio-system',
            'port-forward',
            'deployment/prometheus',
            '%s:9090' % self.port
        ], stdout=subprocess.PIPE)

        self.port_forward.stdout.readline()  # Wait for port forward to be ready

        self.prom = Prometheus('http://localhost:%s/' % self.port)

    def tearDown(self):
        self.port_forward.terminate()

    def test_queries(self):
        for query in test_cases.queries:
            with self.subTest(name=query.description):
                errors = self.prom.run_query(query, debug=True)
                message = 'Alarms Triggered:'
                for e in errors:
                    message += '\n- ' + e
                assert_empty(errors, message)


if __name__ == '__main__':
    unittest.main()

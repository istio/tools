#!/usr/bin/env python3

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

from prometheus import Query, Alarm, Prometheus
import subprocess
import unittest
import os
import signal

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


def setup_promethus():
    port = os.environ.get("PROM_PORT", "9990")
    namespace, deployment = find_prometheus()
    port_forward = subprocess.Popen([
        'kubectl',
        '-n', namespace,
        'port-forward',
        deployment,
        '%s:9090' % port
    ], stdout=subprocess.PIPE)
    port_forward.stdout.readline()  # Wait for port forward to be ready
    return Prometheus('http://localhost:%s/' % port, pid=port_forward.pid)


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
        Query(
            '%s: Envoy CPU' % namespace,
            'rate(container_cpu_usage_seconds_total{container_name="istio-proxy", namespace=~"%s"}[1m]) * %f' % (
                namespace, CPU_MILLI),
            Alarm(
                lambda cpu: cpu > cpu_lim,
                'Envoy CPU is unexpectedly high.'
            ),
            None
        ),
        Query(
            '%s: Envoy Memory' % namespace,
            'max(max_over_time(container_memory_usage_bytes{container_name="istio-proxy", namespace=~"%s"}[1m])) * %f' % (
                namespace, MEM_MB),
            Alarm(
                lambda mem: mem > mem_lim,
                'Envoy memory is unexpectedly high.'
            ),
            None
        ),
        # TODO find a way to get average over time, otherwise this will be flakey and miss real issues.
        # Query(
        #     '%s: CDS Convergence' % namespace,
        #     'count(count_values("value", envoy_cluster_manager_cds_version{namespace="%s"}))' % namespace,
        #     Alarm(
        #         lambda activeVersions: activeVersions > 1,
        #         'CDS has multiple versions running'
        #     )
        # ),
    ]


def istio_requests_sanity(namespace):
    """Ensure that there are some requests to the namespace as a sanity check.
    This won't work for tests which don't report requests through Istio."""
    return Query(
        '%s: Total Requests/s (sanity check)' %
        namespace,
        'sum(rate(istio_requests_total{destination_service_namespace="%s"}[10m]))' %
        namespace,
        Alarm(
            lambda qps: qps < 0.5,
            'There were no requests, the test is likely not running properly.'),
        None)


def stability_query(source, test):
    total = 'sum(rate(stability_outgoing_requests_total{source="%s"}[5m]))' % source
    failure = 'sum(rate(stability_outgoing_requests_total{source="%s", succeeded="False"}[5m]))' % source
    query = Query(
        '{}: error rate'.format(test),
        '{}/{}'.format(failure, total),
        Alarm(
            lambda errs: errs > 0,
            'Error rate too high, expected no errors'
        ),
        'sum(stability_test_instances{test="%s"})' % test
    )
    return query


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

    def test_graceful_shutdown(self):
        queries = [
            *standard_queries('istio-stability-graceful-shutdown'),
            istio_requests_sanity('istio-stability-graceful-shutdown')
        ]
        self.run_queries(queries)

    def test_http_10(self):
        queries = [
            *standard_queries('istio-stability-http10'),
            istio_requests_sanity('istio-stability-http10')
        ]
        self.run_queries(queries)

    def test_mysql(self):
        queries = [
            # TODO get clientside metrics
            *standard_queries('istio-stability-mysql')
        ]
        self.run_queries(queries)

    def test_load_test(self):
        queries = [
            *standard_queries('service-graph..', cpu_lim=250, mem_lim=100)
        ]
        self.run_queries(queries)

    def test_redis(self):
        queries = [
            stability_query(source='redis-client', test='redis')
        ]
        self.run_queries(queries)

    def test_rabbitmq(self):
        queries = [
            stability_query(source='rabbitmq-client', test='rabbitmq')
        ]
        self.run_queries(queries)

    @classmethod
    def setUpClass(self):
        self.prom = setup_promethus()

    @classmethod
    def tearDownClass(self):
        os.kill(self.prom.pid, signal.SIGKILL)

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

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

from prometheus_client import start_http_server, Counter, Gauge
import logging

REQUESTS = Counter(
    'stability_outgoing_requests',
    'Number of requests from this service.',
    ['source', 'destination', 'succeeded']
)


RUNNING = Gauge(
    'stability_test_instances',
    'Is this test running',
    ['test']
)


def report_metrics():
    start_http_server(8080)


def report_running(test):
    RUNNING.labels(test).set_function(lambda: 1)


def attempt_request(f, source, destination, valid=None):
    try:
        response = f()
        if not valid or valid(response):
            succeeded = True
        else:
            succeeded = False
            logging.error(
                "Request from {} to {} had invalid response: {}".format(
                    source, destination, response))

        REQUESTS.labels(source, destination, succeeded).inc()
        return response, succeeded
    except BaseException:
        logging.exception("Request from {} to {} had an exception".format(
            source,
            destination
        ))
        REQUESTS.labels(source, destination, False).inc()
        return None, False

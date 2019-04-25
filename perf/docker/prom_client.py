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
            logging.error("Request from {} to {} had invalid response: {}".format(
                source,
                destination,
                response
            ))

        REQUESTS.labels(source, destination, succeeded).inc()
        return response, succeeded
    except:
        logging.exception("Request from {} to {} had an exception".format(
            source,
            destination
        ))
        REQUESTS.labels(source, destination, False).inc()
        return None, False

import os
import time
import logging
import prom_client
import pika
import sys

password = os.environ["PASSWORD"]
username = os.environ["USERNAME"]
address = os.environ["ADDRESS"]

queue = 'queue'


def with_metrics(f, valid=None):
    return prom_client.attempt_request(
        f,
        source='rabbitmq-client',
        destination='rabbitmq',
        valid=valid
    )


def with_metrics_or_fail(f, valid=None):
    r, success = with_metrics(f, valid)
    if not success:
        raise Exception("Function failed")
    return r, success


def setup_client():
    credentials = pika.PlainCredentials(username, password)
    connection = pika.BlockingConnection(
        pika.ConnectionParameters(address, credentials=credentials))
    channel = connection.channel()
    channel.queue_declare(queue=queue)
    return channel


def send(channel, message):
    with_metrics_or_fail(
        lambda: channel.basic_publish(
            exchange='',
            routing_key=queue,
            body=message
        ),
        valid=None
    )


def attempt_decode(s):
    if s is None:
        return ""
    return s.decode('utf-8')


def receive(channel, expected):
    with_metrics_or_fail(
        lambda: attempt_decode(
            next(channel.consume(queue, inactivity_timeout=1))[2]),
        valid=lambda resp: resp == expected
    )


def run_test():
    pub, succeeded = with_metrics(setup_client)
    if not succeeded:
        logging.error("Failed to setup client")
        sys.exit(1)
    sub, succeeded = with_metrics(setup_client)
    if not succeeded:
        logging.error("Failed to setup client")
        sys.exit(1)

    while True:
        message = "a message"
        send(pub, message)
        receive(sub, message)
        time.sleep(.5)


if __name__ == "__main__":
    prom_client.report_metrics()
    prom_client.report_running('rabbitmq')

    time.sleep(10)  # Wait for server

    while True:
        try:
            run_test()
        except Exception:
            logging.warning("Rerunning test due to exception")
            time.sleep(.5)

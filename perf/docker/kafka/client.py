# Copyright 2019 Istio Authors
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.

import os
import time
import logging
import prom_client
import sys
from kafka import KafkaConsumer, KafkaProducer
from kafka.errors import TopicAlreadyExistsError
from kafka.admin import KafkaAdminClient, NewTopic

address = os.environ["ADDRESS"]

topic = 'stability'


def with_metrics(f, valid=None):
    return prom_client.attempt_request(
        f,
        source='kafka-client',
        destination='kafka',
        valid=valid
    )


def send(queue, message):
    with_metrics(
        lambda: queue.send(
            topic=topic,
            value=bytes(message, encoding="utf-8")
        ),
        valid=lambda resp: resp.get(timeout=1)
    )


def receive(queue, expected):
    with_metrics(
        lambda: next(queue),
        valid=lambda resp: resp.value == expected
    )


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    logging.info("Starting prometheus")
    prom_client.report_metrics()
    prom_client.report_running('kafka')

    logging.info("Setting up topic")
    admin_client, succeeded = with_metrics(
        lambda: KafkaAdminClient(bootstrap_servers=address))
    if not succeeded:
        logging.error("Failed to setup publisher client")
        sys.exit(1)
    try:
        admin_client.create_topics(
            [NewTopic(name=topic, num_partitions=1, replication_factor=1)],
            timeout_ms=1000
        )
        logging.info("Topic created")
    except TopicAlreadyExistsError:
        logging.info("Topic already exists")

    pub, succeeded = with_metrics(
        lambda: KafkaProducer(bootstrap_servers=address))
    if not succeeded:
        logging.error("Failed to setup publisher client")
        sys.exit(1)
    logging.info("Created pub")
    sub, succeeded = with_metrics(lambda: KafkaConsumer(
        bootstrap_servers=address,
        value_deserializer=lambda m: m.decode('utf-8'),
        consumer_timeout_ms=1000,
    ))
    if not succeeded:
        logging.error("Failed to setup subscriber client")
        sys.exit(1)
    logging.info("Created sub")
    sub.subscribe([topic])
    logging.info("Subscribed to topic")

    while True:
        message = "a message"
        logging.info("Sending message")
        send(pub, message)
        logging.info("Reading message")
        receive(sub, message)
        time.sleep(.5)

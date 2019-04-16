import os
import redis
import time
import logging
import prom_client


def setup_redis():
    logging.info("Starting redis master")
    master = redis.Redis(
        host=os.environ['ADDRESS'],
        port=os.environ['PORT'],
        password=os.environ['PASSWORD']
    )

    logging.info("Starting redis slave")
    slave = redis.Redis(
        host=os.environ['SLAVE_ADDRESS'],
        port=os.environ['SLAVE_PORT'],
        password=os.environ['PASSWORD']
    )

    return master, slave


def make_requests(master, slave):
    now = str(time.time())

    prom_client.attempt_request(
        lambda: master.set('time', now),
        'redis-client',
        'redis-master'
    )

    def get_time_from_redis(client, name):
        return prom_client.attempt_request(
            lambda: client.get('time').decode('utf-8'),
            source='redis-client',
            destination=name,
            valid=lambda t: t == now
        )

    get_time_from_redis(master, 'redis-master')
    get_time_from_redis(master, 'redis-slave')


if __name__ == "__main__":
    master, slave = setup_redis()

    prom_client.report_metrics()
    prom_client.report_running('redis')

    while True:
        make_requests(master, slave)
        time.sleep(0.5)

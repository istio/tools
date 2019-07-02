import requests
import collections
from typing import List
import os
import signal

Query = collections.namedtuple(
    'Query',
    ['description', 'query', 'alarm', 'running_query']
)

Alarm = collections.namedtuple(
    'Alarm',
    ['in_alarm', 'error_message']
)


class Prometheus(object):
    def __init__(self, url: str, host: str = None, pid=0):
        self.pid = pid
        self.url = url
        self.headers = {}
        if host is not None:
            self.headers['Host'] = host

    def __del__(self):
        os.kill(self.pid, signal.SIGKILL)

    def fetch_by_query(self, query: str) -> dict:
        resp = requests.get(self.url + '/api/v1/query', {
            'query': query,
        }, headers=self.headers)

        if not resp.ok:
            raise Exception(str(resp))

        return resp.json()

    def fetch_value(self, query: str, default: int = 0) -> float:
        """Runs a query against prometheus and returns the value of the first result.

        TODO: the first result is always used. Future use cases may need to change this.
        """
        resp = self.fetch_by_query(query)['data']['result']
        if not resp:
            return default
        return float(resp[0]['value'][1])

    def run_query(self, query: Query, debug: bool = False) -> List[str]:
        errors = []
        r = self.fetch_value(query.query)
        if query.alarm.in_alarm(r):
            errors.append('{} Response: {}'.format(
                query.alarm.error_message, r))
        if debug:
            print('Testing: %s. Result: %f.' % (query.description, r))
        return errors

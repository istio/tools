#!/usr/bin/env python3
# this program checks the config push latency for the pilot.
import sys
import os
import time
import typing
import subprocess
import argparse

cwd = os.getcwd()
path = os.path.abspath(os.path.join(os.path.dirname( __file__ ), '../../../metrics'))
sys.path.insert(0, path)

from prometheus import Query, Alarm, Prometheus
import check_metrics


# TODO: does this consider namespace?
def envoy_cds_version_count(prom: Prometheus):
    return prom.fetch_value('count(count_values("value", envoy_cluster_manager_cds_version))')


def setup_pilot_loadtest(instance, svc_entry: int):
    helm = 'serviceEntries=%d,instances=%d' % (svc_entry, instance)
    print('setup the loads, %s' % helm)
    env = os.environ
    env['HELM_FLAGS'] = helm
    p = subprocess.Popen([
        './setup.sh',
    ], env=env)
    p.wait()


def wait_till_converge(prom: Prometheus):
    '''Confirm all the Envoy config has been converged to a single version.'''
    occurrence = 0
    while True:
        count = envoy_cds_version_count(prom)
        if count == 1:
            occurrence += 1
        else:
            occurrence = 0
        print('envoy version count %d, occurrences with version_count = 1, occurrences = %d' %(count, occurrence))
        if occurrence == 2:
            return
        time.sleep(3)


def testall(start, end):
    prom = check_metrics.setup_promethus()
    print('finished promethus setup', prom.url)
    setup_pilot_loadtest(start[0],start[1])
    # ensure version is converged.
    wait_till_converge(prom)
    setup_pilot_loadtest(end[0], end[1])
    start = time.time()
    wait_till_converge(prom)
    print('version converged in %s seconds ' % (time.time() - start))

def init_parser():
    parser = argparse.ArgumentParser(
        description='Program for load test.')
    parser.add_argument('-s', '--start',
        nargs=2, type=int,
        default=[1000,200],
        help='initial number of the services and service entries')
    parser.add_argument(
        '-e', '--end',
        nargs=2, type=int,
        default=[1000,205],
        help='the number of the services and service entries to trigger the push')
    return parser.parse_args()

if __name__ == '__main__':
    result = init_parser()
    testall(result.start, result.end)

#!/usr/bin/env python3
# this program checks the config push latency for the pilot.
import sys
import os
import time
import typing
import subprocess

cwd = os.getcwd()
path = os.path.abspath(os.path.join(os.path.dirname( __file__ ), '../../../metrics'))
sys.path.insert(0, path)

from prometheus import Query, Alarm, Prometheus
import check_metrics


def envoy_cds_version_count(prom: Prometheus):
    return prom.fetch_value('count(count_values("value", envoy_cluster_manager_cds_version))')


def setup_pilot_loadtest(instance, svc_entry: int):
    helm = 'serviceEntries=%d,instances=%d' % (svc_entry, instance)
    env = os.environ
    env['HELM_FLAGS'] = helm
    p = subprocess.Popen([
        './setup.sh',
    ], env=env)
    p.wait()


def wait_till_converge(prom: Prometheus):
    while True:
        count = envoy_cds_version_count(prom)
        if count == 1:
            print('envoy version converged, returning')
            return
        print('envoy version not converged yet, count = %s' % count)
        time.sleep(3)


def testall():
    prom = check_metrics.setup_promethus()
    print('finished promethus setup', prom.url)
    setup_pilot_loadtest(100,100)
    # ensure version is converged first.
    wait_till_converge(prom)
    setup_pilot_loadtest(100,110)
    start = time.time()
    wait_till_converge(prom)
    print('version converged in %s seconds ' % (time.time() - start))

if __name__ == '__main__':
    testall()

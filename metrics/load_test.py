#!/usr/bin/env python3
# this program checks the config push latency for the pilot.
from prometheus import Query, Alarm, Prometheus
import subprocess
import unittest
import os
import time
import typing
import check_metrics


def setup_promethus():
    port = os.environ.get("PROM_PORT", "9990")
    namespace, deployment = check_metrics.find_prometheus()
    port_forward = subprocess.Popen([
        'kubectl',
        '-n', namespace,
        'port-forward',
        deployment,
        '%s:9090' % port
    ], stdout=subprocess.PIPE)
    port_forward.stdout.readline()  # Wait for port forward to be ready
    return Prometheus('http://localhost:%s/' % port)


def envoy_cds_version_count(prom: Prometheus):
    return prom.fetch_value('count(count_values("value", envoy_cluster_manager_cds_version))')


def setup_pilot_loadtest(instance, svc_entry: int):
    f = open('pilot-load.yaml', 'w')
    p = subprocess.Popen([
        '/usr/local/bin/helm',
        '--namespace',
        'pilot-load',
        '--set',
        'serviceEntries=%d' % svc_entry,
        '--set',
        'instances=%d' % instance,
        'template',
        '../perf/load/pilot/',
    ], stdout=f)
    p.wait()
    p = subprocess.Popen(
        ['kubectl', 'apply', '-f', 'pilot-load.yaml', '-n', 'pilot-load'])
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
    prom = setup_promethus()
    print('finished promethus setup', prom.url)
    setup_pilot_loadtest(5,5)
    # # ensure version is converged first.
    # wait_till_converge(prom)
    # setup_pilot_loadtest(50,60)
    # start = time.time()
    # wait_till_converge(prom)
    # print('version converged in %s seconds ' % (time.time() - start))

if __name__ == '__main__':
    testall()

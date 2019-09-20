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
#
# this program checks the config push latency for the pilot.
import check_metrics
from prometheus import Query, Alarm, Prometheus
import sys
import os
import time
import typing
import subprocess
import argparse

cwd = os.getcwd()
path = os.path.abspath(os.path.join(os.path.dirname(__file__), '../../../metrics'))
sys.path.insert(0, path)


# count(envoy_cluster_upstream_cx_total{cluster_name="outbound|890||svc-0.pilot-load.svc.cluster.local"})
# envoy_cluster_manager_cds_version is not reliable due to region/zone is not consistently populated.
def config_push_converge_query(prom: Prometheus, svc: str = "svc-0", namespace: str = 'pilot-load'):
    cluster_name = 'outbound|890||{0}.{1}.svc.cluster.local'.format(
        svc, namespace
    )
    result = prom.fetch_by_query(
        'count(envoy_cluster_upstream_cx_total{cluster_name=~".*pilot-load.*"}) by (cluster_name)')
    if not result:
        return []
    return [(point['metric'], point['value'][1])
            for point in result['data']['result']]


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
    start = time.time()
    while True:
        count = config_push_converge_query(prom)
        print('[Query] ', int(time.time() - start), 'seconds, ', count)
        time.sleep(5)


def testall(svc: int, se: int):
    prom = check_metrics.setup_promethus()
    print('finished promethus setup', prom.url)
    setup_pilot_loadtest(svc, se)
    # ensure version is converged.
    wait_till_converge(prom)
    print('version converged in %s seconds ' % (time.time() - start))


def init_parser():
    parser = argparse.ArgumentParser(
        description='Program for load test.')
    parser.add_argument('-s', '--start',
                        nargs=2, type=int,
                        default=[1000, 0],
                        help='initial number of the services and service entries')
    return parser.parse_args()


if __name__ == '__main__':
    result = init_parser()
    testall(result.start[0], result.start[1])

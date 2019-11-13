#!/usr/bin/env python3

# Copyright Istio Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


import sys
import os
import random
import time
import typing
import subprocess
import argparse

import http.server
from urllib.parse import urlparse, parse_qs


TEST_NAMESPACE = 'automtls'
ISTIO_DEPLOY = 'svc-0-automtls-sidecar'
LEGACY_DEPLOY = 'svc-0-automtls-nosidecar'


class testHTTPServer_RequestHandler(http.server.BaseHTTPRequestHandler):

  def do_GET(self):
    self.send_response(200)
    self.send_header('Content-type','text/html')
    self.end_headers()
    query = parse_qs(urlparse(self.path).query)
    istio_percent = random.random()
    if 'istio' in query:
        istio_percent = float(query['istio'][0])
    message = simulate_sidecar_rollout(istio_percent)
    self.wfile.write(bytes(message, "utf8"))
    return


def get_deployment_replicas(namespace, deployment: str):
    cmd = 'kubectl get deployment {dep} -n{ns} {jsonpath}'.format(
        ns=namespace, dep=deployment, jsonpath='''-ojsonpath={.status.replicas}''')
    p = subprocess.Popen(cmd.split(' '), stdout=subprocess.PIPE)
    output = p.communicate()[0]
    if len(output) == 0:
        return 0
    return int(output)


def wait_deployment(namespace, deployment: str):
    cmd = ('kubectl rollout status deployments/{dep} -n{ns}').format(
        dep=deployment,
        ns=namespace
    )
    print(cmd)
    p = subprocess.Popen(cmd.split(' '))
    p.wait()


def scale_deployment(namespace, deployment: str, replica: int):
    cmd = 'kubectl scale deployment {dep} -n{ns} --replicas {replica}'.format(
        dep=deployment, ns=namespace, replica=replica
    )
    print(cmd)
    p = subprocess.Popen(cmd.split(' '))
    p.wait()


def simulate_sidecar_rollout(istio_percent : float):
    '''
    Updates deployments with or without Envoy sidecar.
    wait indicates whether the command wait till all pods become ready.
    '''
    output = 'Namespace {}, sidecar deployment: {}, nosidecar deployment: {}'.format(
        TEST_NAMESPACE, ISTIO_DEPLOY, LEGACY_DEPLOY)
    istio_count = get_deployment_replicas(TEST_NAMESPACE, ISTIO_DEPLOY)
    legacy_count = get_deployment_replicas(TEST_NAMESPACE, LEGACY_DEPLOY)
    total = istio_count + legacy_count
    output = 'sidecar replica {}, legacy replica {}\n\n'.format(
        istio_count, legacy_count)
    istio_count = int(istio_percent * total)
    legacy_count = total - istio_count
    output += ('======================================\n'
            'Scale Istio count {sc}, legacy count {nsc}\n\n').format(
                sc=istio_count, nsc=legacy_count
            )
    scale_deployment(TEST_NAMESPACE, ISTIO_DEPLOY, istio_count)
    scale_deployment(TEST_NAMESPACE, LEGACY_DEPLOY, legacy_count)
    return output


if __name__ == '__main__':
    print('starting the rollout server simulation...')
    server_address = ('127.0.0.1', 8000)
    httpd = http.server.HTTPServer(server_address, testHTTPServer_RequestHandler)
    httpd.serve_forever()

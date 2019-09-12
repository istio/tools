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


import sys
import os
import time
import typing
import subprocess
import argparse

# TODO: is this the pythonic way to add customized path?
sys.path.append('../../../metrics/')

import check_metrics
from prometheus import Query, Alarm, Prometheus

def init_mesh():
  print('Intialize the mesh...')
  p = subprocess.Popen([
      './setup.sh',
  ])
  p.wait()


def deploy_workloads(server_sidecar, server_nosidecar, wait=False):
  '''
  Updates deployments with or without Envoy sidecar.
  wait indicates whether the command wait till all pods become ready.
  '''
  cmd = ('kubectl scale deployment.v1.apps/httpbin-sidecar --replicas={} -nauto-mtls').format(
    server_sidecar)
  p = subprocess.Popen(cmd.split(' '))
  print(cmd)
  p.wait()
  cmd = ('kubectl scale deployment.v1.apps/httpbin-nosidecar --replicas={} -nauto-mtls').format(
    server_nosidecar)
  p = subprocess.Popen(cmd.split(' '))
  p.wait()



def runtest():
  init_mesh()
  deploy_workloads(3, 1)
  time.sleep(20)
  deploy_workloads(1,3)
  time.sleep(20)


if __name__ == "__main__":
  runtest()

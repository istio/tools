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

# Prbobabily of all workloads are migrated to sidecar by scaling.
ALL_SIDECAR_PROB = 0.2
# Change to larger values for real test.
ROLLOUT_INTERVAL_SEC = 5


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


def scale_deployment(namespace, deployment: str, replica : int):
  cmd = 'kubectl scale deployment {dep} -n{ns} --replicas {replica}'.format(
    dep=deployment, ns=namespace, replica=replica
  )
  print(cmd)
  p = subprocess.Popen(cmd.split(' '))
  p.wait()


def simulate_sidecar_rollout(namespace, sidecar_dep, nosidecar_dep: str):
  '''
  Updates deployments with or without Envoy sidecar.
  wait indicates whether the command wait till all pods become ready.
  '''
  print('Namespace {}, sidecar deployment: {}, nosidecar deployment: {}'.format(
      namespace, sidecar_dep, nosidecar_dep))
  sidecar_count = get_deployment_replicas(namespace, sidecar_dep)
  nosidecar_count = get_deployment_replicas(namespace, nosidecar_dep)
  total = sidecar_count + nosidecar_count
  print('sidecar replica {}, nosidecar replica {}'.format(sidecar_count, nosidecar_count))
  iteration = 1
  while True:
    prob = random.random()
    if prob < ALL_SIDECAR_PROB:
      sidecar_count = total
      nosidecar_count = 0
    else:
      sidecar_count = int(random.random() * total)
      nosidecar_count = total - sidecar_count
    print('======================================\n'
      'Scale iteration {itr}, sidecar count {sc}, nosidecar count {nsc}\n\n'.format(
      itr=iteration, sc=sidecar_count, nsc=nosidecar_count
    ))
    scale_deployment(namespace, sidecar_dep, sidecar_count)
    scale_deployment(namespace, nosidecar_dep, nosidecar_count)
    wait_deployment(namespace, sidecar_dep)
    wait_deployment(namespace, nosidecar_dep)
    iteration += 1
    print('\n\n')
    time.sleep(ROLLOUT_INTERVAL_SEC) # random amount of time.


parser = argparse.ArgumentParser(description='Auto mTLS test runner')
parser.add_argument('--sidecar-name', default='svc-0-automtls-sidecar', type=str)
parser.add_argument('--nosidecar-name', default='svc-0-automtls-nosidecar', type=str)
parser.add_argument('-n', '--namespace', default='auto-mtls', type=str)
args = parser.parse_args()


if __name__ == '__main__':
  simulate_sidecar_rollout(args.namespace, args.sidecar_name, args.nosidecar_name)

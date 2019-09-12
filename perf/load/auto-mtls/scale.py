#!/usr/bin/env python3

import sys
import os
import time
import typing
import subprocess
import argparse


def get_deployment_replicas(namespace, deployment: str):
  cmd = 'kubectl get deployment {dep} -n{ns} {jsonpath}'.format(
    ns=namespace, dep=deployment, jsonpath='''-ojsonpath={.status.replicas}''')
  p = subprocess.Popen(cmd.split(' '), stdout=subprocess.PIPE)
  return int(p.communicate()[0])


def wait_deployment(namespace, deployment: str):
  cmd = ('kubectl rollout status deployments/{dep} -n{ns}').format(
    dep=deployment,
    ns=namespace
  )
  print(cmd)
  p = subprocess.Popen(cmd.split(' '))
  p.wait()


def scale_deployment(namespace, deployment: str, replica : int):
  pass


def simulate_sidecar_rollout(namespace, sidecar, nosidecar: str):
  '''
  Updates deployments with or without Envoy sidecar.
  wait indicates whether the command wait till all pods become ready.
  '''
  print('Namespace {}, sidecar deployment: {}, nosidecar deployment: {}'.format(
      namespace, sidecar, nosidecar))
  sidecar_count = get_deployment_replicas(namespace, sidecar)
  nosidecar_count = get_deployment_replicas(namespace, nosidecar)
  print('sidecar replica {}, nosidecar replica {}'.format(sidecar_count, nosidecar_count))
  while True:
    wait_deployment(namespace, sidecar)
    wait_deployment(namespace, nosidecar)
    scale_deployment(namespace, sidecar, 0)
    scale_deployment(namespace, nosidecar, 0)
    time.sleep(3) # random amount of time.


parser = argparse.ArgumentParser(description='Auto mTLS test runner')
parser.add_argument('--sidecar-name', default='svc-0-automtls-sidecar', type=str)
parser.add_argument('--nosidecar-name', default='svc-0-automtls-nosidecar', type=str)
parser.add_argument('-n', '--namespace', default='auto-mtls', type=str)
args = parser.parse_args()


if __name__ == '__main__':
  simulate_sidecar_rollout(args.namespace, args.sidecar_name, args.nosidecar_name)

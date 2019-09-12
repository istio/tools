#!/usr/bin/env python3

import sys
import os
import time
import typing
import subprocess
import argparse


def simulate_sidecar_rollout(sidecar, nosidecar: str):
  '''
  Updates deployments with or without Envoy sidecar.
  wait indicates whether the command wait till all pods become ready.
  '''
  # cmd = ('kubectl scale deployment.v1.apps/httpbin-sidecar --replicas={} -nauto-mtls').format(
    # server_sidecar)
  while True:
    print('Scale the deployment, sidecar deployment: %s, nosidecar deployment: %s', sidecar, nosidecar)
    # cmd = ('kubectl get po -nauto-mtls')
    # p = subprocess.Popen(cmd.split(' '))
    # print(cmd)
    # p.wait()
    time.sleep(3)
  # cmd = ('kubectl scale deployment.v1.apps/httpbin-nosidecar --replicas={} -nauto-mtls').format(
    # server_nosidecar)
  # p = subprocess.Popen(cmd.split(' '))
  # p.wait()

parser = argparse.ArgumentParser(description='Auto mTLS test runner')
parser.add_argument('--sidecar-name', default='svc-0-automtls-sidecar', type=str)
parser.add_argument('--nosidecar-name', default='svc-0-automtls-nosidecar', type=str)
args = parser.parse_args()


if __name__ == '__main__':
  simulate_sidecar_rollout(args.sidecar_name, args.nosidecar_name)

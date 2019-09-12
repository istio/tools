#!/usr/bin/env python3

import sys
import os
import time
import typing
import subprocess
import argparse


def simulate_sidecar_rollout(server_sidecar, server_nosidecar, wait=False):
  '''
  Updates deployments with or without Envoy sidecar.
  wait indicates whether the command wait till all pods become ready.
  '''
  # cmd = ('kubectl scale deployment.v1.apps/httpbin-sidecar --replicas={} -nauto-mtls').format(
    # server_sidecar)
  while True:
    # cmd = ('kubectl get po -nauto-mtls')
    # p = subprocess.Popen(cmd.split(' '))
    # print(cmd)
    # p.wait()
    print('jianfeih is debugging')
    print("your message", file=sys.stderr)
    time.sleep(3)
  # cmd = ('kubectl scale deployment.v1.apps/httpbin-nosidecar --replicas={} -nauto-mtls').format(
    # server_nosidecar)
  # p = subprocess.Popen(cmd.split(' '))
  # p.wait()

if __name__ == '__main__':
  simulate_sidecar_rollout()

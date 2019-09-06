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

# this program checks the config push latency for the pilot.
import check_metrics
from prometheus import Query, Alarm, Prometheus
import sys
import os
import time
import typing
import subprocess
import argparse


def init_mesh():
  pass


def deploy_workloads(server_sidecar, server_nosidecar, wait=False):
  '''
  Updates deployments with or without Envoy sidecar.
  wait indicates whether the command wait till all pods become ready.
  '''
  pass


def start_requets():
  pass


def stop_requests():
  pass


def gather_metrics():
  pass


def runtest():
  init_mesh()
  start_requets()
  stop_requests()
  gather_metrics()


if __name__ == "__main__":
  runtest()

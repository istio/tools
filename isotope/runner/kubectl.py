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

"""Abstractions for common calls to kubectl."""

import contextlib
import logging
import socket
import subprocess
import tempfile
import time
from typing import Any, Dict, Generator, List

import yaml

from . import sh


@contextlib.contextmanager
def manifest(path: str, cleanup=False) -> Generator[None, None, None]:
    """Runs `kubectl apply -f path` on entry and opposing delete on exit."""
    try:
        apply_file(path)
        yield
    finally:
        if cleanup:
            delete_file(path)


def apply_file(path: str) -> None:
    sh.run_kubectl(['apply', '-f', path], check=True)


def delete_file(path: str) -> None:
    sh.run_kubectl(['delete', '-f', path])


def apply_dicts(dicts: List[Dict[str, Any]],
                intermediate_file_path: str = None) -> None:
    yaml_str = yaml.dump_all(dicts)
    apply_text(yaml_str, intermediate_file_path=intermediate_file_path)


def apply_text(json_or_yaml: str, intermediate_file_path: str = None) -> None:
    """Creates/updates resources described in either JSON or YAML string.

    Uses `kubectl apply -f FILE`.

    Args:
        json_or_yaml: contains either the JSON or YAML manifest of the
                resource(s) to apply; applied through an intermediate file
        intermediate_file_path: if set, defines the file to write to (useful
                for debugging); otherwise, uses a temporary file
    """
    if intermediate_file_path is None:
        opener = tempfile.NamedTemporaryFile(mode='w+')
    else:
        opener = open(intermediate_file_path, 'w+')

    with opener as f:
        f.write(json_or_yaml)
        f.flush()
        apply_file(f.name)


@contextlib.contextmanager
def port_forward(label_key: str, label_value: str, target_port: int,
                 namespace: str) -> Generator[int, None, None]:
    """Port forwards the first pod matching label, yielding the open port."""
    # TODO: Catch error if label matches zero pods.
    pod_name = sh.run_kubectl(
        [
            'get', 'pod', '-l{}={}'.format(label_key, label_value),
            '-o=jsonpath={.items[0].metadata.name}', '--namespace', namespace
        ],
        check=True).stdout
    local_port = _get_open_port()
    proc = subprocess.Popen(
        [
            'kubectl', '--namespace', namespace, 'port-forward', pod_name,
            '{}:{}'.format(local_port, target_port)
        ],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE)

    try:
        # proc.communicate waits until the process terminates or timeout.
        _, stderr_bytes = proc.communicate(timeout=1)

        # If proc terminates after 1 second, assume that an error occured.
        stderr = stderr_bytes.decode('utf-8') if stderr_bytes else ''
        info = ': {}'.format(stderr) if stderr else ''
        msg = 'could not port-forward to {}:{} on local port {}{}'.format(
            pod_name, target_port, local_port, info)
        raise RuntimeError(msg)
    except subprocess.TimeoutExpired:
        # If proc is still running after 1 second, assume that proc will
        # continue port forwarding until termination, as expected.
        pass

    yield local_port

    proc.terminate()


# Adapted from
# https://stackoverflow.com/questions/2838244/get-open-tcp-port-in-python.
def _get_open_port() -> int:
    sock = socket.socket()
    sock.bind(('', 0))
    _, port = sock.getsockname()
    return port

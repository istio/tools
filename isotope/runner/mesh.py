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

"""Definitions for abstracting a mesh environment underlying a topology.

Any new environments should add an entry to the if-else chain in `for_state`
and define three functions:

1. set_up: creates the necessary Kubernetes resources
2. get_ingress_url: returns the URL to access the ingress or entrypoint
3. tear_down: deletes the previously created resources

These three functions make it simple for use in a with-statement.
"""
from __future__ import print_function

import contextlib
from typing import Callable, Generator

from . import config, consts, istio as istio_lib


class Environment:
    """Bundles functions to set up, tear down, and interface with a mesh."""

    def __init__(self, name: str, set_up: Callable[[], None],
                 tear_down: Callable[[], None],
                 get_ingress_url: Callable[[], str]) -> None:
        self.name = name
        self.set_up = set_up
        self.tear_down = tear_down
        self.get_ingress_url = get_ingress_url

    @contextlib.contextmanager
    def context(self) -> Generator[str, None, None]:
        try:
            self.set_up()
            yield self.get_ingress_url()
        finally:
            self.tear_down()


def none(entrypoint_service_name: str, entrypoint_service_port: int,
         entrypoint_service_namespace: str) -> Environment:
    def get_ingress_url() -> str:
        return 'http://{}.{}.svc.cluster.local:{}'.format(
            entrypoint_service_name, entrypoint_service_namespace,
            entrypoint_service_port)

    return Environment(
        name='none',
        set_up=_do_nothing,
        tear_down=_do_nothing,
        get_ingress_url=get_ingress_url)


def istio(entrypoint_service_name: str, entrypoint_service_namespace: str,
          archive_url: str, values: str, tear_down=False) -> Environment:
    def set_up() -> None:
        istio_lib.set_up(entrypoint_service_name, entrypoint_service_namespace,
                         archive_url, values)

    td = _do_nothing
    if tear_down:
        td = istio_lib.tear_down
    return Environment(
        name='istio',
        set_up=set_up,
        tear_down=td,
        get_ingress_url=istio_lib.get_ingress_gateway_url)


def for_state(name: str, entrypoint_service_name: str,
              entrypoint_service_namespace: str,
              config: config.RunnerConfig, values: str) -> Environment:
    if name == 'NONE':
        env = none(entrypoint_service_name, consts.SERVICE_PORT,
                   consts.SERVICE_GRAPH_NAMESPACE)
    elif name == 'ISTIO':
        env = istio(entrypoint_service_name, entrypoint_service_namespace,
                    config.istio_archive_url, values)
    else:
        raise ValueError('{} is not a known environment'.format(name))

    return env


def _do_nothing():
    print("empty teardown")
    pass

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

"""Defines constant paths for static and generated YAML."""

import os

_RESOURCES_DIR = os.path.realpath(
    os.path.join(os.getcwd(), os.path.dirname(__file__)))

HELM_SERVICE_ACCOUNT_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                              'helm-service-account.yaml')
PERSISTENT_VOLUME_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'persistent-volume.yaml')

PROMETHEUS_VALUES_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                               'values-prometheus.gen.yaml')
SERVICE_GRAPH_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                           'service-graph.gen.yaml')
ISTIO_GEN_YAML_PATH = os.path.join(_RESOURCES_DIR, 'istio.gen.yaml')
ISTIO_INGRESS_YAML_PATH = os.path.join(_RESOURCES_DIR,
                                       'istio-ingress.gen.yaml')

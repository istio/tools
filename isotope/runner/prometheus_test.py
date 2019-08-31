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

import yaml

from . import prometheus


def test_values_should_return_correct_yaml():
    expected = {
        'prometheus': {
            'storageSpec': {
                'volumeClaimTemplate': {
                    'spec': {
                        'accessModes': ['ReadWriteOnce'],
                        'resources': {
                            'requests': {
                                'storage': '10G'
                            }
                        },
                        'storageClassName': '',
                        'volumeName': 'prometheus-persistent-volume'
                    }
                }
            },
            'serviceMonitors': [{
                'endpoints': [{
                    'metricRelabelings': [{
                        'replacement': 'tjberry',
                        'targetLabel': 'user'
                    }, {
                        'replacement': 'stuff',
                        'targetLabel': 'custom'
                    }],
                    'targetPort':
                    8080
                }],
                'namespaceSelector': {
                    'matchNames': ['service-graph']
                },
                'selector': {
                    'matchLabels': {
                        'app': 'service-graph'
                    }
                },
                'name':
                'service-graph-monitor'
            }, {
                'endpoints': [{
                    'metricRelabelings': [{
                        'replacement': 'tjberry',
                        'targetLabel': 'user'
                    }, {
                        'replacement': 'stuff',
                        'targetLabel': 'custom'
                    }],
                    'targetPort':
                    42422
                }],
                'namespaceSelector': {
                    'matchNames': ['default']
                },
                'selector': {
                    'matchLabels': {
                        'app': 'client'
                    }
                },
                'name':
                'client-monitor'
            }, {
                'endpoints': [{
                    'metricRelabelings': [{
                        'replacement': 'tjberry',
                        'targetLabel': 'user'
                    }, {
                        'replacement': 'stuff',
                        'targetLabel': 'custom'
                    }],
                    'targetPort':
                    42422
                }],
                'namespaceSelector': {
                    'matchNames': ['istio-system']
                },
                'selector': {
                    'matchLabels': {
                        'istio': 'mixer'
                    }
                },
                'name':
                'istio-mixer-monitor'
            }]
        },
        'deployExporterNode': True,
        'deployAlertManager': False,
        'deployKubeState': True,
        'deployKubeDNS': True,
        'deployKubeScheduler': True,
        'deployKubelets': True,
        'deployGrafana': True,
        'deployKubeControllerManager': True,
        'deployKubeEtcd': True,
        'exporter-kubelets': {
            'https': False
        }
    }

    labels = {
        'user': 'tjberry',
        'custom': 'stuff',
    }

    actual = prometheus._get_values(labels)

    assert expected == actual

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

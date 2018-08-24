"""Functions for creating and updating the Prometheus configuration."""

import logging
import time
import tempfile
import textwrap
from typing import Any, Dict, List

import yaml

from . import consts, sh, wait

_HELM_RELEASE_NAME = 'kube-prometheus'


def apply(labels: Dict[str, str] = {},
          intermediate_file_path: str = None) -> None:
    """Creates or updates Prometheus values to add labels to all metrics.

    Args:
        json_or_yaml: contains either the JSON or YAML manifest of the
                resource(s) to apply; applied through an intermediate file
        intermediate_file_path: if set, defines the file to write to (useful
                for debugging); otherwise, uses a temporary file
    """
    logging.info('applying Prometheus configuration')

    config = _get_values(labels)
    values_yaml = yaml.dump(config, default_flow_style=False)

    if intermediate_file_path is None:
        opener = tempfile.NamedTemporaryFile(mode='w+')
    else:
        opener = open(intermediate_file_path, 'w+')

    with opener as f:
        f.write(values_yaml)
        f.flush()

        _apply_prometheus_values(f.name)


def _apply_prometheus_values(path: str) -> None:
    proc = sh.run_with_k8s_api(['helm', 'get', _HELM_RELEASE_NAME])
    already_exists = proc.returncode == 0
    if already_exists:
        _update_prometheus(path)
    else:
        _install_prometheus(path)


def _update_prometheus(values_path: str) -> None:
    logging.debug('updating coreos/kube-prometheus')
    sh.run_with_k8s_api(
        [
            'helm', 'upgrade', _HELM_RELEASE_NAME, 'coreos/kube-prometheus',
            '--values', values_path
        ],
        check=True)
    # TODO: Should wait until Prometheus is actually updated.
    time.sleep(5 * 60)


def _install_prometheus(values_path: str) -> None:
    logging.debug('installing coreos/kube-prometheus')
    sh.run_with_k8s_api(
        [
            'helm', 'install', 'coreos/kube-prometheus', '--name',
            _HELM_RELEASE_NAME, '--namespace', consts.MONITORING_NAMESPACE,
            '--values', values_path
        ],
        check=True)
    wait.until_stateful_sets_are_ready(consts.MONITORING_NAMESPACE)


def _get_values(labels: Dict[str, str]) -> Dict[str, Any]:
    return {
        'deployAlertManager': False,
        'deployExporterNode': True,
        'deployGrafana': True,
        'deployKubeControllerManager': True,
        'deployKubeDNS': True,
        'deployKubeEtcd': True,
        'deployKubelets': True,
        'deployKubeScheduler': True,
        'deployKubeState': True,
        'exporter-kubelets': {
            # Must be false for GKE.
            'https': False,
        },
        'prometheus': _get_prometheus_config(labels)
    }


def _get_prometheus_config(labels: Dict[str, str]) -> Dict[str, Any]:
    metric_relabelings = _get_metric_relabelings(labels)
    return {
        'serviceMonitors': [
            _get_service_monitor('service-graph-monitor', 8080,
                                 consts.SERVICE_GRAPH_NAMESPACE,
                                 {'app': 'service-graph'}, metric_relabelings),
            _get_service_monitor('client-monitor', 42422,
                                 consts.DEFAULT_NAMESPACE, {'app': 'client'},
                                 metric_relabelings),
            _get_service_monitor('istio-mixer-monitor', 42422,
                                 consts.ISTIO_NAMESPACE, {'istio': 'mixer'},
                                 metric_relabelings),
        ],
        'storageSpec':
        _get_storage_spec(),
    }


def _get_service_monitor(
        name: str, port: int, namespace: str, match_labels: Dict[str, str],
        metric_relabelings: List[Dict[str, Any]]) -> Dict[str, Any]:
    return {
        'name':
        name,
        'endpoints': [{
            'targetPort': port,
            'metricRelabelings': metric_relabelings,
        }],
        'namespaceSelector': {
            'matchNames': [namespace],
        },
        'selector': {
            'matchLabels': match_labels,
        },
    }


def _get_metric_relabelings(labels: Dict[str, str]) -> List[Dict[str, Any]]:
    return [{
        'targetLabel': key,
        'replacement': value,
    } for key, value in labels.items()]


def _get_storage_spec() -> Dict[str, Any]:
    return {
        'volumeClaimTemplate': {
            'spec': {
                'accessModes': ['ReadWriteOnce'],
                'resources': {
                    'requests': {
                        'storage': '10G',
                    },
                },
                'volumeName': 'prometheus-persistent-volume',
                'storageClassName': '',
            },
        },
    }

"""Functions for manipulating the Istio environment."""

import contextlib
import logging
import os
import tarfile
import tempfile
from typing import Any, Dict, Generator

import requests
import yaml

from . import consts, kubectl, resources, sh, wait


def set_up(entrypoint_service_name: str, entrypoint_service_namespace: str,
           archive_url: str) -> None:
    """Installs Istio from the archive URL.

    Requires Helm client to be present.

    This downloads and extracts the archive in a temporary directory, then
    installs the resources via `helm template` and `kubectl apply`.
    """
    with tempfile.TemporaryDirectory() as tmp_dir_path:
        archive_path = os.path.join(tmp_dir_path, 'istio.tar.gz')
        _download(archive_url, archive_path)

        extracted_dir_path = os.path.join(tmp_dir_path, 'istio')
        extracted_istio_path = _extract(archive_path, extracted_dir_path)

        chart_path = os.path.join(extracted_istio_path, 'install',
                                  'kubernetes', 'helm', 'istio')
        _install(
            chart_path,
            consts.ISTIO_NAMESPACE,
            intermediate_file_path=resources.ISTIO_GEN_YAML_PATH)

        _create_ingress_rules(entrypoint_service_name,
                              entrypoint_service_namespace)


def get_ingress_gateway_url() -> str:
    ip = wait.until_output([
        'kubectl', '--namespace', consts.ISTIO_NAMESPACE, 'get', 'service',
        'istio-ingressgateway', '-o',
        'jsonpath={.status.loadBalancer.ingress[0].ip}'
    ])
    return 'http://{}:{}'.format(ip, consts.ISTIO_INGRESS_GATEWAY_PORT)


# Adapted from
# http://docs.python-requests.org/en/latest/user/advanced/#body-content-workflow.
def _download(archive_url: str, path: str) -> None:
    logging.info('downloading %s', archive_url)
    with requests.get(archive_url, stream=True) as response:
        with open(path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=1024):
                if chunk:  # Filter out keep-alive new chunks.
                    f.write(chunk)


def _extract(archive_path: str, extracted_dir_path: str) -> str:
    """Extracts the .tar.gz at archive_path to extracted_dir_path.

    Args:
        archive_path: path to a .tar.gz archive file, containing a single
                directory when extracted
        extracted_dir_path: the destination in which to extract the contents
                of the archive

    Returns:
        the path to the single directory the archive contains
    """
    with tarfile.open(archive_path) as tar:
        tar.extractall(path=extracted_dir_path)
    extracted_items = os.listdir(extracted_dir_path)
    if len(extracted_items) != 1:
        raise ValueError(
            'archive at {} did not contain a single directory'.format(
                archive_path))
    return os.path.join(extracted_dir_path, extracted_items[0])


def _install(chart_path: str, namespace: str,
             intermediate_file_path: str) -> None:
    logging.info('installing Helm chart for Istio')
    sh.run_kubectl(['create', 'namespace', namespace])
    # TODO: Why is it necessary to set the hub and tag when these are already
    # in the chart?
    istio_yaml = sh.run(
        [
            'helm', 'template', chart_path, '--namespace', namespace,
            '--set=global.hub=docker.io/istionightly',
            '--set=global.tag=nightly-master',
            '--set=global.proxy.resources.requests.cpu=1000m',
            '--set=global.proxy.resources.limits.cpu=1000m',
            '--set=global.proxy.resources.requests.memory=256Mi',
            '--set=global.proxy.resources.limits.memory=256Mi',
            '--set=global.defaultResources.requests.cpu=1000m',
            '--set=global.defaultResources.limits.cpu=1000m'
        ],
        check=True).stdout
    kubectl.apply_text(
        istio_yaml, intermediate_file_path=intermediate_file_path)
    wait.until_deployments_are_ready(namespace)


@contextlib.contextmanager
def _work_dir(path: str) -> Generator[None, None, None]:
    prev_path = os.getcwd()
    if not os.path.exists(path):
        os.makedirs(path)
    os.chdir(path)
    yield
    os.chdir(prev_path)


def _create_ingress_rules(entrypoint_service_name: str,
                          entrypoint_service_namespace: str) -> None:
    logging.info('creating istio ingress rules')
    ingress_yaml = _get_ingress_yaml(entrypoint_service_name,
                                     entrypoint_service_namespace)
    kubectl.apply_text(
        ingress_yaml, intermediate_file_path=resources.ISTIO_INGRESS_YAML_PATH)


def _get_ingress_yaml(entrypoint_service_name: str,
                      entrypoint_service_namespace: str) -> str:
    gateway = _get_gateway_dict()
    virtual_service = _get_virtual_service_dict(entrypoint_service_name,
                                                entrypoint_service_namespace)
    return yaml.dump_all([gateway, virtual_service], default_flow_style=False)


def _get_gateway_dict() -> Dict[str, Any]:
    return {
        'apiVersion': 'networking.istio.io/v1alpha3',
        'kind': 'Gateway',
        'metadata': {
            'name': 'entrypoint-gateway',
        },
        'spec': {
            'selector': {
                'istio': 'ingressgateway',
            },
            'servers': [{
                'hosts': ['*'],
                'port': {
                    'name': 'http',
                    'number': consts.ISTIO_INGRESS_GATEWAY_PORT,
                    'protocol': 'HTTP',
                },
            }],
        },
    }


def _get_virtual_service_dict(
        entrypoint_service_name: str,
        entrypoint_service_namespace: str) -> Dict[str, Any]:
    return {
        'apiVersion': 'networking.istio.io/v1alpha3',
        'kind': 'VirtualService',
        'metadata': {
            'name': 'entrypoint',
        },
        'spec': {
            'hosts': ['*'],
            'gateways': ['entrypoint-gateway'],
            'http': [{
                'route': [{
                    'destination': {
                        'host':
                        '{}.{}.svc.cluster.local'.format(
                            entrypoint_service_name,
                            entrypoint_service_namespace),
                        'port': {
                            'number': consts.SERVICE_PORT,
                        },
                    },
                }],
            }],
        },
    }


def tear_down() -> None:
    """Deletes the Istio resources and namespace."""
    sh.run_kubectl(['delete', '-f', resources.ISTIO_GEN_YAML_PATH])
    sh.run_kubectl(['delete', 'namespace', consts.ISTIO_NAMESPACE])
    wait.until_namespace_is_deleted(consts.SERVICE_GRAPH_NAMESPACE)

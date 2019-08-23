"""Functions for manipulating the Istio environment."""
from __future__ import print_function

import contextlib
import logging
import os
import tarfile
import tempfile
from typing import Any, Dict, Generator

import yaml
import time

from . import consts, kubectl, resources, sh, wait

DAILY_BUILD_URL = "https://storage.googleapis.com/istio-prerelease/daily-build"


def convert_archive(archive_url: str) -> str:
    """Convert symbolic archive into archive url

    """
    if archive_url.startswith("http"):
        return archive_url

    full_name = "{}-09-15".format(archive_url)

    return "{daily}/{full_name}/istio-{full_name}-linux.tar.gz".format(
        daily=DAILY_BUILD_URL, full_name=full_name)


def set_up(entrypoint_service_name: str, entrypoint_service_namespace: str,
           archive_url: str, values: str) -> None:
    """Installs Istio from the archive URL.

    Requires Helm client to be present.

    This downloads and extracts the archive in a temporary directory, then
    installs the resources via `helm template` and `kubectl apply`.
    """
    archive_url = convert_archive(archive_url)

    print(("Using archive_url", archive_url))

    with tempfile.TemporaryDirectory() as tmp_dir_path:
        archive_path = os.path.join(tmp_dir_path, 'istio.tar.gz')
        _download(archive_url, archive_path)

        extracted_dir_path = os.path.join(tmp_dir_path, 'istio')
        extracted_istio_path = _extract(archive_path, extracted_dir_path)

        crd_path = os.path.join(extracted_istio_path, 'install',
                                'kubernetes', 'helm', 'istio-init')

        chart_path = os.path.join(extracted_istio_path, 'install',
                                  'kubernetes', 'helm', 'istio')

        _apply_crds(
            crd_path,
            'istio-init',
            consts.ISTIO_NAMESPACE)

        _install(
            chart_path,
            consts.ISTIO_NAMESPACE,
            intermediate_file_path=resources.ISTIO_GEN_YAML_PATH,
            values=values)

        _create_ingress_rules(entrypoint_service_name,
                              entrypoint_service_namespace)


def get_ingress_gateway_url() -> str:
    ip = wait.until_output([
        'kubectl', '--namespace', consts.ISTIO_NAMESPACE, 'get', 'service',
        'istio-ingressgateway', '-o',
        'jsonpath={.status.loadBalancer.ingress[0].ip}'
    ])
    return 'http://{}:{}'.format(ip, consts.ISTIO_INGRESS_GATEWAY_PORT)


def _download(archive_url: str, path: str) -> None:
    logging.info('downloading %s', archive_url)
    sh.run(['curl', '-L', '--output', path, archive_url])


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


def _apply_crds(path: str, name: str, namespace: str) -> None:
    logging.info('applying crd definitions for Istio')
    sh.run_kubectl(['create', 'namespace', namespace])

    istio_yaml = sh.run(
        [
            'helm',
            'template',
            path,
            '--name',
            name,
            '--namespace',
            namespace
        ],
        check=True).stdout
    kubectl.apply_text(istio_yaml)

    logging.info('sleeping for 30 seconds as an extra buffer')
    time.sleep(30)
    wait.until_deployments_are_ready(namespace)


def _install(chart_path: str, namespace: str,
             intermediate_file_path: str, values: str) -> None:
    logging.info('installing Helm chart for Istio')
    istio_yaml = sh.run(
        [
            'helm',
            'template',
            chart_path,
            '--namespace',
            namespace,
            '--values',
            values
            # TODO: Use a values file, specified in the TOML configuration.
            # Consider replacing environments with a list of values files, then
            # each of those values files represents the environment. This code
            # can apply those against the chart.
            # '--set=global.proxy.resources.requests.cpu=1000m',
            # '--set=global.proxy.resources.requests.memory=256Mi',
            # '--set=global.defaultResources.requests.cpu=1000m',
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

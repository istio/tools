"""Defines run which follows the testing pipeline after cluster creation."""

import contextlib
import logging
import os
import time
from typing import Dict, Generator, Optional

import requests

from . import consts, entrypoint, istio, kubectl, md5, mesh, prometheus, \
              resources, sh, wait

_REPO_ROOT = os.path.join(os.getcwd(),
                          os.path.dirname(os.path.dirname(__file__)))
_MAIN_GO_PATH = os.path.join(_REPO_ROOT, 'convert', 'main.go')


def run(topology_path: str, env: mesh.Environment, service_image: str,
        client_image: str, istio_archive_url: str, test_qps: Optional[int],
        test_duration: str, test_num_concurrent_connections: int,
        static_labels: Dict[str, str], deploy_prometheus=False) -> None:
    """Runs a load test on the topology in topology_path with the environment.

    Args:
        topology_path: the path to the file containing the topology
        env: the pre-existing mesh environment for the topology (i.e. Istio)
        service_image: the Docker image to represent each node in the topology
        client_image: the Docker image which can run a load test (i.e. Fortio)
        istio_archive_url: URL to access a released tar.gz archive via an
                HTTP GET request
        test_qps: the target QPS for the client; None = max
        test_duration: the duration for the client to run
        test_num_concurrent_connections: the number of simultaneous connections
                for the client to make
        static_labels: labels to add to each Prometheus monitor
    """

    manifest_path = _gen_yaml(topology_path, service_image,
                              test_num_concurrent_connections, client_image,
                              env.name)

    topology_name = _get_basename_no_ext(topology_path)
    labels = {
        'environment': env.name,
        'topology_name': topology_name,
        'topology_hash': md5.hex(topology_path),
        **static_labels,
    }
    if deploy_prometheus:
      prometheus.apply(
          labels,
          intermediate_file_path=resources.PROMETHEUS_VALUES_GEN_YAML_PATH)

    with env.context() as ingress_url:
        logging.info('starting test with environment "%s"', env.name)
        result_output_path = '{}_{}.json'.format(topology_name, env.name)

        _test_service_graph(manifest_path, result_output_path, ingress_url,
                            test_qps, test_duration,
                            test_num_concurrent_connections)


def _get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def _gen_yaml(topology_path: str, service_image: str,
              max_idle_connections_per_host: int, client_image: str,
              env_name: str) -> str:
    """Converts topology_path to Kubernetes manifests.

    The neighboring Go command in convert/ handles this operation.

    Args:
        topology_path: the path containing the topology YAML
        service_image: the Docker image to represent each node in the topology;
                passed to the Go command
        client_image: the Docker image which can run a load test (i.e. Fortio);
                passed to the Go command
        env_name: the environment name (i.e. "NONE" or "ISTIO")
    """
    logging.info('generating Kubernetes manifests from %s', topology_path)
    service_graph_node_selector = _get_gke_node_selector(
        consts.SERVICE_GRAPH_NODE_POOL_NAME)
    client_node_selector = _get_gke_node_selector(consts.CLIENT_NODE_POOL_NAME)
    sh.run(
        [
            'go', 'run', _MAIN_GO_PATH, 'kubernetes', '--service-image',
            service_image, '--service-max-idle-connections-per-host',
            str(max_idle_connections_per_host), '--client-image', client_image,
            topology_path, resources.SERVICE_GRAPH_GEN_YAML_PATH,
            service_graph_node_selector, client_node_selector,
            "--environment-name", env_name
        ],
        check=True)
    return resources.SERVICE_GRAPH_GEN_YAML_PATH


def _get_gke_node_selector(node_pool_name: str) -> str:
    return 'cloud.google.com/gke-nodepool={}'.format(node_pool_name)


def _test_service_graph(yaml_path: str, test_result_output_path: str,
                        test_target_url: str, test_qps: Optional[int],
                        test_duration: str,
                        test_num_concurrent_connections: int) -> None:
    """Deploys the service graph at yaml_path and runs a load test on it."""
    # TODO: extract to env.context, with entrypoint hostname as the ingress URL
    with kubectl.manifest(yaml_path):
        wait.until_deployments_are_ready(consts.SERVICE_GRAPH_NAMESPACE)
        wait.until_service_graph_is_ready()
        # TODO: Why is this extra buffer necessary?
        logging.debug('sleeping for 30 seconds as an extra buffer')
        time.sleep(30)

        _run_load_test(test_result_output_path, test_target_url, test_qps,
                       test_duration, test_num_concurrent_connections)

        wait.until_prometheus_has_scraped()

    wait.until_namespace_is_deleted(consts.SERVICE_GRAPH_NAMESPACE)


def _run_load_test(result_output_path: str, test_target_url: str,
                   test_qps: Optional[int], test_duration: str,
                   test_num_concurrent_connections: int) -> None:
    """Sends an HTTP request to the client; expecting a JSON response.

    The HTTP request's query string contains the necessary info to perform
    the load test, adapted from the arguments described in
    https://github.com/istio/istio/blob/master/tools/README.md#run-the-functions.

    Args:
        result_output_path: the path to write the JSON output.
        test_target_url: the in-cluster URL to
        test_qps: the target QPS for the client; None = max
        test_duration: the duration for the client to run
        test_num_concurrent_connections: the number of simultaneous connections
                for the client to make
    """
    logging.info('starting load test')
    with kubectl.port_forward("app", consts.CLIENT_NAME, consts.CLIENT_PORT,
                              consts.DEFAULT_NAMESPACE) as local_port:
        qps = -1 if test_qps is None else test_qps  # -1 indicates max QPS.
        url = ('http://localhost:{}/fortio'
               '?json=on&qps={}&t={}&c={}&load=Start&url={}').format(
                   local_port, qps, test_duration,
                   test_num_concurrent_connections, test_target_url)
        result = _http_get_json(url)
    _write_to_file(result_output_path, result)


def _http_get_json(url: str) -> str:
    """Sends an HTTP GET request to url, returning its JSON response."""
    response = None
    while response is None:
        try:
            response = requests.get(url)
        except (requests.ConnectionError, requests.HTTPError) as e:
            logging.error('%s; retrying request to %s', e, url)
    return response.text


def _write_to_file(path: str, contents: str) -> None:
    logging.debug('writing contents to %s', path)
    with open(path, 'w') as f:
        f.writelines(contents)

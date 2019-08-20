"""Defines run which follows the testing pipeline after cluster creation."""

import contextlib
import logging
import os
import time
import yaml
from typing import Dict, Generator, Optional, List
import json

import requests

from . import consts, entrypoint, istio, kubectl, md5, mesh, prometheus, \
    resources, sh, wait

_REPO_ROOT = os.path.join(os.getcwd(),
                          os.path.dirname(os.path.dirname(__file__)))
_MAIN_GO_PATH = os.path.join(_REPO_ROOT, 'convert', 'main.go')


def run(topology_path: str,
        env: mesh.Environment,
        service_image: str,
        client_image: str,
        istio_archive_url: str,
        policy_files: List[str],
        test_qps: List[Optional[int]],
        test_duration: List[str],
        test_num_concurrent_connections: List[int],
        client_attempts: int,
        static_labels: Dict[str, str],
        policy_dir: str,
        deploy_prometheus=False) -> None:
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
    if service_image == None:
        manifest_path = _gen_load_gen_yaml(client_image, policy_dir)
    else:
        manifest_path = _gen_yaml(topology_path, service_image,
                                  test_num_concurrent_connections,
                                  client_image, env.name)

    if service_image == None:
        topology_name = None
        topology_hash = None
    else:
        topology_name = _get_basename_no_ext(topology_path)
        topology_hash = md5.hex(topology_path)

    labels = {
        'environment': env.name,
        'topology_name': topology_name,
        'topology_hash': topology_hash,
        **static_labels,
    }
    if deploy_prometheus:
        prometheus.apply(
            labels,
            intermediate_file_path=resources.PROMETHEUS_VALUES_GEN_YAML_PATH)

    with env.context() as ingress_urls:
        logging.info('starting test with environment "%s"', env.name)
        result_output_path = '{}_{}'.format(topology_name, env.name)

        if service_image == None:
            actual_app = True
        else:
            actual_app = False

        _test_service_graph(env, manifest_path, result_output_path,
                            ingress_urls, test_qps, test_duration,
                            test_num_concurrent_connections, client_attempts,
                            actual_app, policy_dir)


def _apply_policy_files(policy_dir: str, namespace: str) -> None:
    logging.info('applying policy files')

    for file in os.listdir(policy_dir):
        file_path = (os.path.join(policy_dir, file))
        kubectl.apply_file(file_path)


def _get_basename_no_ext(path: str) -> str:
    basename = os.path.basename(path)
    return os.path.splitext(basename)[0]


def _gen_load_gen_yaml(client_image: str, policy_dir: str):
    """Generates Kubernetes manifests about fortio client

    The neighboring Go command in convert/ handles this operation.

    Args:
        client_image: the Docker image which can run a load test (i.e. Fortio);
                passed to the Go command
        env_name: the environment name (i.e. "NONE" or "ISTIO")
    """
    logging.info('generating Load Generator manifests from')
    client_node_selector = _get_gke_node_selector(consts.CLIENT_NODE_POOL_NAME)
    gen = sh.run([
        'go',
        'run',
        _MAIN_GO_PATH,
        'fortio',
        '--client-image',
        client_image,
        '--client-node-selector',
        client_node_selector,
        policy_dir,
    ],
                 check=True)
    with open(resources.SERVICE_GRAPH_GEN_YAML_PATH, 'w') as f:
        f.write(gen.stdout)

    return resources.SERVICE_GRAPH_GEN_YAML_PATH


def _gen_yaml(topology_path: str, service_image: str,
              max_idle_connections_per_host: List[int], client_image: str,
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
    gen = sh.run([
        'go',
        'run',
        _MAIN_GO_PATH,
        'kubernetes',
        '--service-image',
        service_image,
        '--service-max-idle-connections-per-host',
        str(max(max_idle_connections_per_host)),
        '--client-image',
        client_image,
        "--environment-name",
        env_name,
        '--service-node-selector',
        service_graph_node_selector,
        '--client-node-selector',
        client_node_selector,
        '--load-level',
        str(consts.INITIAL_LOAD_LEVEL),
        topology_path,
    ],
                 check=True)
    with open(resources.SERVICE_GRAPH_GEN_YAML_PATH, 'w') as f:
        f.write(gen.stdout)

    return resources.SERVICE_GRAPH_GEN_YAML_PATH


def _get_gke_node_selector(node_pool_name: str) -> str:
    return 'cloud.google.com/gke-nodepool={}'.format(node_pool_name)


def _test_service_graph(
        env: mesh.Environment, yaml_path: str, test_result_output_path: str,
        test_target_urls: List[str], test_qps: List[Optional[int]],
        test_duration: List[str], test_num_concurrent_connections: List[int],
        client_attempts: str, actual_app: bool, policy_dir: str) -> None:
    """Deploys the service graph at yaml_path and runs a load test on it."""
    # TODO: extract to env.context, with entrypoint hostname as the ingress URL
    try:
        with kubectl.manifest(yaml_path):
            wait.until_deployments_are_ready(consts.SERVICE_GRAPH_NAMESPACE)
            if not actual_app:
                wait.until_service_graph_is_ready()

            if env.name == "istio":
                _apply_policy_files(policy_dir, consts.ISTIO_NAMESPACE)

            # TODO: Why is this extra buffer necessary?
            logging.debug('sleeping for 120 seconds as an extra buffer')
            time.sleep(120)

            _run_load_test(test_result_output_path, test_target_urls, test_qps,
                           test_duration, test_num_concurrent_connections,
                           client_attempts, actual_app, policy_dir)

            wait.until_prometheus_has_scraped()
    except Exception as e:
        print("Error: ", e)
    finally:
        sh.run_kubectl(['delete', 'ns', consts.SERVICE_GRAPH_NAMESPACE],
                       check=True)
        wait.until_namespace_is_deleted(consts.SERVICE_GRAPH_NAMESPACE)

def _run_load_test(result_output_path: str, test_target_urls: List[str],
                   test_qps: List[Optional[int]], test_duration: List[str],
                   test_num_concurrent_connections: List[int],
                   client_attempts: int, actual_app: bool,
                   policy_dir: str) -> None:
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
    for qps in test_qps:
        for num_connections in test_num_concurrent_connections:
            for duration in test_duration:
                for test_target_url in test_target_urls:
                    for attempt in range(client_attempts):
                        logging.info(
                            'starting load test, [QPS: %s,' + \
                            'Number Connections: %s, Duration: %s,' + \
                            'URL: %s, Policies: %s, Attempt: %s]',
                            str(qps), str(num_connections),
                            duration, test_target_url, policy_dir,
                            str(attempt))

                        with kubectl.port_forward(
                                "app", consts.CLIENT_NAME, consts.CLIENT_PORT,
                                consts.DEFAULT_NAMESPACE) as local_port:
                            qps = -1 if qps is None else qps  # -1 indicates max QPS.
                            percentiles = ','.join([
                                str(percentile)
                                for percentile in range(1, 100)
                            ])
                            url = (
                                'http://localhost:{}/fortio'
                                '?json=on&qps={}&t={}&c={}&load=Start&p={}&url={}'
                            ).format(local_port, qps, duration,
                                     num_connections, percentiles,
                                     test_target_url)
                            result = _http_get_json(url)

                        output_path = result_output_path + "_" + str(
                            qps) + "_" + str(num_connections) + "_" + str(
                                duration)

                        path = policy_dir.replace('/', '_')
                        output_path += "_" + path + "_" + str(
                            attempt) + ".json"

                        _write_to_file(output_path, result)

                        _run_tratis(
                            json.loads(result)["StartTime"], (duration),
                            output_path)

                        logging.info("Sleeping for 60 seconds")
                        time.sleep(60)


def _run_tratis(start_time: str, duration: str, file_name: str) -> None:
    traces_file = "TRACES_" + file_name
    results_file = "RESULTS_" + file_name

    command = [
        "go", "run", "main.go", "-start", start_time, "-duration", duration,
        "-tool", "jaeger", "-traces", traces_file, "-results", results_file
    ]

    sh.run(command, cwd="../tratis/service/", check=True)


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

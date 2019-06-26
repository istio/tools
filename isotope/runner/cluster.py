"""Set up for GKE clusters and Prometheus monitoring."""

import logging
import os

from . import consts, prometheus, resources, sh, wait


def set_up_if_not_exists(
        project_id: str, name: str, zone: str, version: str,
        service_graph_machine_type: str, service_graph_disk_size_gb: int,
        service_graph_num_nodes: int, client_machine_type: str,
        client_disk_size_gb: int) -> None:
    sh.run_gcloud(['config', 'set', 'project', project_id], check=True)

    # TODO: This is the default tabular output. Filter the input to just the
    # names of the existing clusters.
    output = sh.run_gcloud(
        ['container', 'clusters', 'list', '--zone', zone], check=True).stdout
    # TODO: Also check if the cluster is normal (e.g. not being deleted).
    if name in output:
        logging.debug('%s already exists; bypassing creation', name)
    else:
        logging.debug('%s does not exist yet; creating...', name)
        set_up(project_id, name, zone, version, service_graph_machine_type,
               service_graph_disk_size_gb, service_graph_num_nodes,
               client_machine_type, client_disk_size_gb)


def set_up(project_id: str, name: str, zone: str, version: str,
           service_graph_machine_type: str, service_graph_disk_size_gb: int,
           service_graph_num_nodes: int, client_machine_type: str,
           client_disk_size_gb: int, deploy_prometheus=False) -> None:
    """Creates and sets up a GKE cluster.

    Args:
        project_id: full ID for the cluster's GCP project
        name: name of the GKE cluster
        zone: GCE zone (e.g. "us-central1-a")
        version: GKE version (e.g. "1.9.7-gke.3")
        service_graph_machine_type: GCE type of service machines
        service_graph_disk_size_gb: disk size of service machines in gigabytes
        service_graph_num_nodes: number of machines in the service graph pool
        client_machine_type: GCE type of client machine
        client_disk_size_gb: disk size of client machine in gigabytes
    """
    sh.run_gcloud(['config', 'set', 'project', project_id], check=True)

    _create_cluster(name, zone, version, 'n1-standard-4', 16, 1)
    _create_cluster_role_binding()

    if deploy_prometheus:
        _create_persistent_volume()
        _initialize_helm()
        _helm_add_prometheus_operator()
        prometheus.apply(
            intermediate_file_path=resources.PROMETHEUS_VALUES_GEN_YAML_PATH)

    _create_service_graph_node_pool(service_graph_num_nodes,
                                    service_graph_machine_type,
                                    service_graph_disk_size_gb,
                                    zone)
    _create_client_node_pool(client_machine_type, client_disk_size_gb, zone)


def _create_cluster(name: str, zone: str, version: str, machine_type: str,
                    disk_size_gb: int, num_nodes: int) -> None:
    logging.info('creating cluster "%s"', name)
    sh.run_gcloud(
        [
            'container', 'clusters', 'create', name, '--zone', zone,
            '--cluster-version', version, '--machine-type', machine_type,
            '--disk-size',
            str(disk_size_gb), '--num-nodes',
            str(num_nodes)
        ],
        check=True)
    sh.run_gcloud(['config', 'set', 'container/cluster', name], check=True)
    sh.run_gcloud(
        ['container', 'clusters', 'get-credentials', '--zone', zone, name],
        check=True)


def _create_service_graph_node_pool(num_nodes: int, machine_type: str,
                                    disk_size_gb: int, zone: str) -> None:
    logging.info('creating service graph node-pool')
    _create_node_pool(consts.SERVICE_GRAPH_NODE_POOL_NAME, num_nodes,
                      machine_type, disk_size_gb, zone)


def _create_client_node_pool(machine_type: str, disk_size_gb: int,
                             zone: str) -> None:
    logging.info('creating client node-pool')
    _create_node_pool(consts.CLIENT_NODE_POOL_NAME, 1, machine_type,
                      disk_size_gb, zone)


def _create_node_pool(name: str, num_nodes: int, machine_type: str,
                      disk_size_gb: int, zone: str) -> None:
    sh.run_gcloud(
        [
            'container', 'node-pools', 'create', name, '--machine-type',
            machine_type, '--num-nodes',
            str(num_nodes), '--disk-size',
            str(disk_size_gb), '--zone',
            zone
        ],
        check=True)


def _create_cluster_role_binding() -> None:
    logging.info('creating cluster-admin-binding')
    proc = sh.run_gcloud(['config', 'get-value', 'account'], check=True)
    account = proc.stdout
    sh.run_kubectl(
        [
            'create', 'clusterrolebinding', 'cluster-admin-binding',
            '--clusterrole', 'cluster-admin', '--user', account
        ],
        check=True)


def _create_persistent_volume() -> None:
    logging.info('creating persistent volume')
    sh.run_kubectl(
        ['apply', '-f', resources.PERSISTENT_VOLUME_YAML_PATH], check=True)


def _initialize_helm() -> None:
    logging.info('initializing Helm')
    sh.run_kubectl(
        ['create', '-f', resources.HELM_SERVICE_ACCOUNT_YAML_PATH], check=True)
    sh.run_with_k8s_api(
        ['helm', 'init', '--service-account', 'tiller', '--wait'], check=True)
    sh.run_with_k8s_api(
        [
            'helm', 'repo', 'add', 'coreos',
            'https://s3-eu-west-1.amazonaws.com/coreos-charts/stable'
        ],
        check=True)


def _helm_add_prometheus_operator() -> None:
    logging.info('installing coreos/prometheus-operator')
    sh.run_with_k8s_api(
        [
            'helm', 'install', 'coreos/prometheus-operator', '--name',
            'prometheus-operator', '--namespace', consts.MONITORING_NAMESPACE
        ],
        check=True)

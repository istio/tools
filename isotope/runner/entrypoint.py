"""Read topology YAML and extract information about the entrypoint service."""

import yaml

from . import consts


def extract_name(topology_path: str) -> str:
    """Returns the name of the entrypoint service in the topology."""
    with open(topology_path, 'r') as f:
        topology = yaml.load(f)

    services = topology['services']
    entrypoint_services = [svc for svc in services if svc.get('isEntrypoint')]
    if len(entrypoint_services) != 1:
        raise ValueError(
            'topology at {} should only have one entrypoint'.format(
                topology_path))
    entrypoint_name = entrypoint_services[0]['name']
    return entrypoint_name


def extract_url(topology_path: str) -> str:
    """Returns the in-cluster URL to access the service graph's entrypoint."""
    entrypoint_name = extract_name(topology_path)
    url = 'http://{}.{}.svc.cluster.local:{}'.format(
        entrypoint_name, consts.SERVICE_GRAPH_NAMESPACE, consts.SERVICE_PORT)
    return url

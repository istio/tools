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

"""Read runner configuration from a dict or TOML."""

from typing import Any, Dict, List, Optional

import toml


class RunnerConfig:
    """Represents the intermediary between a config file"""

    def __init__(self, topology_paths: List[str], environments: List[str],
                 istio_archive_url: str, cluster_project_id: str,
                 cluster_name: str, cluster_zones: List[str],
                 cluster_version: str, server_machine_type: str,
                 server_disk_size_gb: int, server_num_nodes: int,
                 server_image: str, client_machine_type: str,
                 client_disk_size_gb: int, client_num_nodes: int,
                 client_image: str, client_qps: Optional[int],
                 client_duration: str, client_num_conc_conns: int) -> None:
        self.topology_paths = topology_paths
        self.environments = environments
        self.istio_archive_url = istio_archive_url
        self.cluster_project_id = cluster_project_id
        self.cluster_name = cluster_name
        self.cluster_zones = cluster_zones
        self.cluster_version = cluster_version
        self.server_machine_type = server_machine_type
        self.server_disk_size_gb = server_disk_size_gb
        self.server_num_nodes = server_num_nodes
        self.server_image = server_image
        self.client_machine_type = client_machine_type
        self.client_disk_size_gb = client_disk_size_gb
        self.client_num_nodes = client_num_nodes
        self.client_image = client_image
        self.client_qps = client_qps
        self.client_duration = client_duration
        self.client_num_conc_conns = client_num_conc_conns

    def labels(self) -> Dict[str, str]:
        """Returns the static labels for Prometheus for this configuration."""
        return {
            'istio_archive_url': self.istio_archive_url,
            'cluster_version': self.cluster_version,
            'cluster_zones': self.cluster_zones,
            'server_machine_type': self.server_machine_type,
            'server_disk_size_gb': str(self.server_disk_size_gb),
            'server_num_nodes': str(self.server_num_nodes),
            'server_image': self.server_image,
            'client_machine_type': self.client_machine_type,
            'client_disk_size_gb': str(self.client_disk_size_gb),
            'client_num_nodes': str(self.client_num_nodes),
            'client_image': self.client_image,
            'client_qps': str(self.client_qps),
            'client_duration': self.client_duration,
            'client_num_concurrent_connections':
            str(self.client_num_conc_conns),
        }


def from_dict(d: Dict[str, Any]) -> RunnerConfig:
    topology_paths = d.get('topology_paths', [])
    environments = d.get('environments', [])

    istio = d['istio']
    istio_archive_url = istio['archive_url']

    cluster = d['cluster']
    cluster_project_id = cluster['project_id']
    cluster_name = cluster['name']
    cluster_zones = cluster['zones']
    cluster_version = cluster['version']

    server = d['server']
    server_machine_type = server['machine_type']
    server_disk_size_gb = server['disk_size_gb']
    server_num_nodes = server['num_nodes']
    server_image = server['image']

    client = d['client']
    client_machine_type = client['machine_type']
    client_disk_size_gb = client['disk_size_gb']
    client_num_nodes = client['num_nodes']
    client_image = client['image']
    client_qps = client['qps']
    if client_qps == 'max':
        client_qps = None
    else:
        # Must coerce into integer, otherwise not a valid QPS.
        client_qps = int(client_qps)
    client_duration = client['duration']
    client_num_conc_conns = client['num_concurrent_connections']

    return RunnerConfig(
        topology_paths=topology_paths,
        environments=environments,
        istio_archive_url=istio_archive_url,
        cluster_project_id=cluster_project_id,
        cluster_name=cluster_name,
        cluster_zones=cluster_zones,
        cluster_version=cluster_version,
        server_machine_type=server_machine_type,
        server_disk_size_gb=server_disk_size_gb,
        server_image=server_image,
        server_num_nodes=server_num_nodes,
        client_machine_type=client_machine_type,
        client_disk_size_gb=client_disk_size_gb,
        client_image=client_image,
        client_qps=client_qps,
        client_duration=client_duration,
        client_num_nodes=client_num_nodes,
        client_num_conc_conns=client_num_conc_conns)


def from_toml_file(path: str) -> RunnerConfig:
    d = toml.load(path)
    return from_dict(d)

#!/usr/bin/env python3

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

# Creates a complete tree of services (every level is filled).

import collections
from typing import Any, Dict, List

import yaml

REQUEST_SIZE = 128
RESPONSE_SIZE = 128
NUM_REPLICAS = 1

# Depth of the tree.
NUM_LEVELS = 3
# Amount of dependent or child services each service has.
NUM_BRANCHES = 3
NUM_SERVICES = sum([NUM_BRANCHES**i for i in range(NUM_LEVELS)])

Service = Dict[str, Any]


def main() -> None:
    entrypoint = {
        'name': 'svc-0',
        'isEntrypoint': True,
    }
    services_paths = collections.deque([(entrypoint, ['0'])])
    services = []  # type: List[Service]

    for _ in range(NUM_SERVICES):
        current_service, current_path = services_paths.popleft()
        services.append(current_service)
        remaining_services = NUM_SERVICES - len(services) - len(services_paths)
        if remaining_services > 0:
            child_services = []  # type: List[Service]
            for child_service_i in range(
                    min(NUM_BRANCHES, remaining_services)):
                child_path = current_path.copy()
                child_path.append(str(child_service_i))
                child_service_name = 'svc-{}'.format('-'.join(child_path))
                child_service = {
                    'name': child_service_name,
                }  # type: Dict[str, Any]

                child_services.append(child_service)
                services_paths.append((child_service, child_path))

            current_service['script'] = _call_all(child_services)

    with open('gen.yaml', 'w') as f:
        yaml.dump(
            {
                'defaults': {
                    'requestSize': REQUEST_SIZE,
                    'responseSize': RESPONSE_SIZE,
                    'numReplicas': NUM_REPLICAS,
                },
                'services': services,
            },
            f,
            default_flow_style=False)


def _call_all(svcs: List[Service]) -> List[List[Dict]]:
    return [[{'call': svc['name']} for svc in svcs]]


if __name__ == '__main__':
    main()

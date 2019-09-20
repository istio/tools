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

"""Module for automating topology testing.

The pseudo-code for the intended calls for this is:

```
read configuration
create cluster
add prometheus
for each topology:
  convert topology to Kubernetes YAML
  for each environment (none, istio, sidecars only, etc.):
    update Prometheus labels
    deploy environment
    deploy topology
    run load test
    delete topology
    delete environment
delete cluster
```
"""

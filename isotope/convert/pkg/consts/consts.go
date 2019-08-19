// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package consts

const (
	// ServiceContainerName is the name to assign the container when it is run.
	ServiceContainerName = "mock-service"

	// ServicePort is the port the service will run on.
	ServicePort = 8080
	// ServicePortName is the name of the service port.
	ServicePortName = "http-web"

	// ServiceGraphNamespace is the name of the namespace that all service graph
	// related components will live in.
	ServiceGraphNamespace = "service-graph"

	// ConfigPath is the parent directory of all service configuration files.
	ConfigPath = "/etc/config"
	// ServiceGraphYAMLFileName is the name of the file which contains the
	// YAML-unmarshallable ServiceGraph.
	ServiceGraphYAMLFileName = "service-graph.yaml"
	// ServiceGraphConfigMapKey is the key of the Kubernetes config map entry
	// holding the ServiceGraph's YAML to be mounted in
	// "${ConfigPath}/${ServiceGraphYAMLFileName}".
	ServiceGraphConfigMapKey = "service-graph"

	// ServiceNameEnvKey is the key of the environment variable whose value is
	// the name of the service.
	ServiceNameEnvKey = "SERVICE_NAME"

	// Load Level (QPS) of the experiment.
	LoadEnvKey = "LOAD"

	// FortioMetricsPort is the port on which /metrics is available.
	FortioMetricsPort = 42422
)

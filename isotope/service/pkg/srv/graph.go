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

package srv

import (
	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/ghodss/yaml"
	"istio.io/fortio/log"
	"istio.io/tools/isotope/convert/pkg/graph"
	"istio.io/tools/isotope/convert/pkg/graph/size"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

// HandlerFromServiceGraphYAML makes a handler to emulate the service with name
// serviceName in the service graph represented by the YAML file at path.
func HandlerFromServiceGraphYAML(
	path string, serviceName string) (Handler, error) {

	serviceGraph, err := serviceGraphFromYAMLFile(path)
	if err != nil {
		return Handler{}, err
	}

	service, err := extractService(serviceGraph, serviceName)
	if err != nil {
		return Handler{}, err
	}
	logService(service)

	serviceTypes := extractServiceTypes(serviceGraph)

	responsePayload, err := makeRandomByteArray(service.ResponseSize)
	if err != nil {
		return Handler{}, err
	}

	return Handler{
		Service:         service,
		ServiceTypes:    serviceTypes,
		responsePayload: responsePayload,
	}, nil
}

func makeRandomByteArray(n size.ByteSize) ([]byte, error) {
	arr := make([]byte, n)
	if _, err := rand.Read(arr); err != nil {
		return nil, err
	}
	return arr, nil
}

func logService(service svc.Service) error {
	if log.Log(log.Info) {
		serviceYAML, err := yaml.Marshal(service)
		if err != nil {
			return err
		}
		log.Infof("acting as service %s:\n%s", service.Name, serviceYAML)
	}
	return nil
}

// serviceGraphFromYAMLFile unmarshals the ServiceGraph from the YAML at path.
func serviceGraphFromYAMLFile(
	path string) (serviceGraph graph.ServiceGraph, err error) {
	graphYAML, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	log.Debugf("unmarshalling\n%s", graphYAML)
	err = yaml.Unmarshal(graphYAML, &serviceGraph)
	if err != nil {
		return
	}
	return
}

// extractService finds the service in serviceGraph with the specified name.
func extractService(
	serviceGraph graph.ServiceGraph, name string) (
	service svc.Service, err error) {
	for _, svc := range serviceGraph.Services {
		if svc.Name == name {
			service = svc
			return
		}
	}
	err = fmt.Errorf(
		"service with name %s does not exist in %v", name, serviceGraph)
	return
}

// extractServiceTypes builds a map from service name to its type
// (i.e. HTTP or gRPC).
func extractServiceTypes(
	serviceGraph graph.ServiceGraph) map[string]svctype.ServiceType {
	types := make(map[string]svctype.ServiceType, len(serviceGraph.Services))
	for _, service := range serviceGraph.Services {
		types[service.Name] = service.Type
	}
	return types
}

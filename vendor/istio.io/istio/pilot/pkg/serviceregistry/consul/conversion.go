// Copyright 2017 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package consul

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"

	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pkg/log"
)

const (
	protocolTagName = "protocol"
	externalTagName = "external"
)

func convertLabels(labels []string) model.Labels {
	out := make(model.Labels, len(labels))
	for _, tag := range labels {
		vals := strings.Split(tag, "|")
		// Labels not of form "key|value" are ignored to avoid possible collisions
		if len(vals) > 1 {
			out[vals[0]] = vals[1]
		} else {
			log.Warnf("Tag %v ignored since it is not of form key|value", tag)
		}
	}
	return out
}

func convertPort(port int, name string) *model.Port {
	if name == "" {
		name = "tcp"
	}

	return &model.Port{
		Name:     name,
		Port:     port,
		Protocol: convertProtocol(name),
	}
}

func convertService(endpoints []*api.CatalogService) *model.Service {
	name := ""

	meshExternal := false
	resolution := model.ClientSideLB

	ports := make(map[int]*model.Port)
	for _, endpoint := range endpoints {
		name = endpoint.ServiceName

		port := convertPort(endpoint.ServicePort, endpoint.ServiceMeta[protocolTagName])

		if svcPort, exists := ports[port.Port]; exists && svcPort.Protocol != port.Protocol {
			log.Warnf("Service %v has two instances on same port %v but different protocols (%v, %v)",
				name, port.Port, svcPort.Protocol, port.Protocol)
		} else {
			ports[port.Port] = port
		}

		// TODO This will not work if service is a mix of external and local services
		// or if a service has more than one external name
		if endpoint.ServiceMeta[externalTagName] != "" {
			meshExternal = true
			resolution = model.Passthrough
		}
	}

	svcPorts := make(model.PortList, 0, len(ports))
	for _, port := range ports {
		svcPorts = append(svcPorts, port)
	}

	hostname := serviceHostname(name)
	out := &model.Service{
		Hostname:     hostname,
		Address:      "0.0.0.0",
		Ports:        svcPorts,
		MeshExternal: meshExternal,
		Resolution:   resolution,
		Attributes: model.ServiceAttributes{
			Name:      string(hostname),
			Namespace: model.IstioDefaultConfigNamespace,
		},
	}

	return out
}

func convertInstance(instance *api.CatalogService) *model.ServiceInstance {
	labels := convertLabels(instance.ServiceTags)
	port := convertPort(instance.ServicePort, instance.ServiceMeta[protocolTagName])

	addr := instance.ServiceAddress
	if addr == "" {
		addr = instance.Address
	}

	meshExternal := false
	resolution := model.ClientSideLB
	externalName := instance.ServiceMeta[externalTagName]
	if externalName != "" {
		meshExternal = true
		resolution = model.DNSLB
	}

	hostname := serviceHostname(instance.ServiceName)
	return &model.ServiceInstance{
		Endpoint: model.NetworkEndpoint{
			Address:     addr,
			Port:        instance.ServicePort,
			ServicePort: port,
			Locality:    instance.Datacenter,
		},
		Service: &model.Service{
			Hostname:     hostname,
			Address:      instance.ServiceAddress,
			Ports:        model.PortList{port},
			MeshExternal: meshExternal,
			Resolution:   resolution,
			Attributes: model.ServiceAttributes{
				Name:      string(hostname),
				Namespace: model.IstioDefaultConfigNamespace,
			},
		},
		Labels: labels,
	}
}

// serviceHostname produces FQDN for a consul service
func serviceHostname(name string) model.Hostname {
	// TODO include datacenter in Hostname?
	// consul DNS uses "redis.service.us-east-1.consul" -> "[<optional_tag>].<svc>.service.[<optional_datacenter>].consul"
	return model.Hostname(fmt.Sprintf("%s.service.consul", name))
}

// parseHostname extracts service name from the service hostname
func parseHostname(hostname model.Hostname) (name string, err error) {
	parts := strings.Split(string(hostname), ".")
	if len(parts) < 1 || parts[0] == "" {
		err = fmt.Errorf("missing service name from the service hostname %q", hostname)
		return
	}
	name = parts[0]
	return
}

func convertProtocol(name string) model.Protocol {
	protocol := model.ParseProtocol(name)
	if protocol == model.ProtocolUnsupported {
		log.Warnf("unsupported protocol value: %s", name)
		return model.ProtocolTCP
	}
	return protocol
}

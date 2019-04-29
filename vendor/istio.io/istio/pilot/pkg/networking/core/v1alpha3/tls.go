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

package v1alpha3

import (
	"strings"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pilot/pkg/model"
	"istio.io/istio/pilot/pkg/networking/util"
)

// Match by source labels, the listener port where traffic comes in, the gateway on which the rule is being
// bound, etc. All these can be checked statically, since we are generating the configuration for a proxy
// with predefined labels, on a specific port.
func matchTLS(match *v1alpha3.TLSMatchAttributes, proxyLabels model.LabelsCollection, gateways map[string]bool, port int) bool {
	if match == nil {
		return true
	}

	gatewayMatch := len(match.Gateways) == 0
	for _, gateway := range match.Gateways {
		gatewayMatch = gatewayMatch || gateways[gateway]
	}

	labelMatch := proxyLabels.IsSupersetOf(model.Labels(match.SourceLabels))

	portMatch := match.Port == 0 || match.Port == uint32(port)

	return gatewayMatch && labelMatch && portMatch
}

// Match by source labels, the listener port where traffic comes in, the gateway on which the rule is being
// bound, etc. All these can be checked statically, since we are generating the configuration for a proxy
// with predefined labels, on a specific port.
func matchTCP(match *v1alpha3.L4MatchAttributes, proxyLabels model.LabelsCollection, gateways map[string]bool, port int) bool {
	if match == nil {
		return true
	}

	gatewayMatch := len(match.Gateways) == 0
	for _, gateway := range match.Gateways {
		gatewayMatch = gatewayMatch || gateways[gateway]
	}

	labelMatch := proxyLabels.IsSupersetOf(model.Labels(match.SourceLabels))

	portMatch := match.Port == 0 || match.Port == uint32(port)

	return gatewayMatch && labelMatch && portMatch
}

// Select the config pertaining to the service being processed.
func getConfigsForHost(host model.Hostname, configs []model.Config) []model.Config {
	svcConfigs := make([]model.Config, 0)
	for index := range configs {
		virtualService := configs[index].Spec.(*v1alpha3.VirtualService)
		for _, vsHost := range virtualService.Hosts {
			if model.Hostname(vsHost).Matches(host) {
				svcConfigs = append(svcConfigs, configs[index])
				break
			}
		}
	}
	return svcConfigs
}

// hashRuntimeTLSMatchPredicates hashes runtime predicates of a TLS match
func hashRuntimeTLSMatchPredicates(match *v1alpha3.TLSMatchAttributes) string {
	return strings.Join(match.SniHosts, ",") + "|" + strings.Join(match.DestinationSubnets, ",")
}

func buildSidecarOutboundTLSFilterChainOpts(env *model.Environment, node *model.Proxy, push *model.PushContext, destinationCIDR string,
	service *model.Service, listenPort *model.Port, proxyLabels model.LabelsCollection,
	gateways map[string]bool, configs []model.Config) []*filterChainOpts {

	if !listenPort.Protocol.IsTLS() {
		return nil
	}

	// TLS matches are composed of runtime and static predicates.
	// Static predicates can be evaluated during the generation of the config. Examples: gateway, source labels, etc.
	// Runtime predicates cannot be evaluated during config generation. Instead the proxy must be configured to
	// evaluate them. Examples: SNI hosts, source/destination subnets, etc.
	//
	// A list of matches may contain duplicate runtime matches, but different static matches. For example:
	//
	// {sni_hosts: A, sourceLabels: X} => destination M
	// {sni_hosts: A, sourceLabels: *} => destination N
	//
	// For a proxy with labels X, we can evaluate the static predicates to get:
	// {sni_hosts: A} => destination M
	// {sni_hosts: A} => destination N
	//
	// The matches have the same runtime predicates. Since the second match can never be reached, we only
	// want to generate config for the first match.
	//
	// To achieve this in this function we keep track of which runtime matches we have already generated config for
	// and only add config if the we have not already generated config for that set of runtime predicates.
	matchHasBeenHandled := make(map[string]bool) // Runtime predicate set -> have we generated config for this set?

	// Is there a virtual service with a TLS block that matches us?
	hasTLSMatch := false

	out := make([]*filterChainOpts, 0)
	for _, config := range configs {
		virtualService := config.Spec.(*v1alpha3.VirtualService)
		for _, tls := range virtualService.Tls {
			for _, match := range tls.Match {
				if matchTLS(match, proxyLabels, gateways, listenPort.Port) {
					// Use the service's CIDRs.
					// But if a virtual service overrides it with its own destination subnet match
					// give preference to the user provided one
					// destinationCIDR will be empty for services with VIPs
					destinationCIDRs := []string{destinationCIDR}
					// Only set CIDR match if the listener is bound to an IP.
					// If its bound to a unix domain socket, then ignore the CIDR matches
					// Unix domain socket bound ports have Port value set to 0
					if len(match.DestinationSubnets) > 0 && listenPort.Port > 0 {
						destinationCIDRs = match.DestinationSubnets
					}
					matchHash := hashRuntimeTLSMatchPredicates(match)
					if !matchHasBeenHandled[matchHash] {
						out = append(out, &filterChainOpts{
							metadata:         util.BuildConfigInfoMetadata(config.ConfigMeta),
							sniHosts:         match.SniHosts,
							destinationCIDRs: destinationCIDRs,
							networkFilters:   buildOutboundNetworkFilters(env, node, tls.Route, push, listenPort, config.ConfigMeta),
						})
						hasTLSMatch = true
					}
					matchHasBeenHandled[matchHash] = true
				}
			}
		}
	}

	// HTTPS or TLS ports without associated virtual service
	if !hasTLSMatch {
		var sniHosts []string

		// In case of a sidecar config with user defined port, if the user specified port is not the same as the
		// service's port, then pick the service port if and only if the service has only one port. If service
		// has multiple ports, then route to a cluster with the listener port (i.e. sidecar defined port) - the
		// traffic will most likely blackhole.
		port := listenPort.Port
		if len(service.Ports) == 1 {
			port = service.Ports[0].Port
		}

		clusterName := model.BuildSubsetKey(model.TrafficDirectionOutbound, "", service.Hostname, port)
		// Use the hostname as the SNI value if and only if we dont have a destination VIP or if the destination is a CIDR.
		// In both cases, the listener will be bound to 0.0.0.0. So SNI match is the only way to distinguish different
		// target services. If we have a VIP, then we know the destination. There is no need to do a SNI match. It saves us from
		// having to generate expensive permutations of the host name just like RDS does..
		// NOTE that we cannot have two services with the same VIP as our listener build logic will treat it as a collision and
		// ignore one of the services.
		svcListenAddress := service.GetServiceAddressForProxy(node)
		if strings.Contains(svcListenAddress, "/") {
			// Address is a CIDR, already captured by destinationCIDR parameter.
			svcListenAddress = ""
		}

		if len(destinationCIDR) > 0 || len(svcListenAddress) == 0 || svcListenAddress == WildcardAddress {
			sniHosts = []string{string(service.Hostname)}
		}

		out = append(out, &filterChainOpts{
			sniHosts:         sniHosts,
			destinationCIDRs: []string{destinationCIDR},
			networkFilters:   buildOutboundNetworkFiltersWithSingleDestination(env, node, clusterName, listenPort),
		})
	}

	return out
}

func buildSidecarOutboundTCPFilterChainOpts(env *model.Environment, node *model.Proxy, push *model.PushContext, destinationCIDR string,
	service *model.Service, listenPort *model.Port, proxyLabels model.LabelsCollection,
	gateways map[string]bool, configs []model.Config) []*filterChainOpts {

	if listenPort.Protocol.IsTLS() {
		return nil
	}

	out := make([]*filterChainOpts, 0)

	// very basic TCP
	// break as soon as we add one network filter with no destination addresses to match
	// This is the terminating condition in the filter chain match list
	defaultRouteAdded := false
TcpLoop:
	for _, config := range configs {
		virtualService := config.Spec.(*v1alpha3.VirtualService)
		for _, tcp := range virtualService.Tcp {
			destinationCIDRs := []string{destinationCIDR}
			if len(tcp.Match) == 0 {
				// implicit match
				out = append(out, &filterChainOpts{
					metadata:         util.BuildConfigInfoMetadata(config.ConfigMeta),
					destinationCIDRs: destinationCIDRs,
					networkFilters:   buildOutboundNetworkFilters(env, node, tcp.Route, push, listenPort, config.ConfigMeta),
				})
				defaultRouteAdded = true
				break TcpLoop
			}

			// Use the service's virtual address first.
			// But if a virtual service overrides it with its own destination subnet match
			// give preference to the user provided one
			virtualServiceDestinationSubnets := make([]string, 0)

			for _, match := range tcp.Match {
				if matchTCP(match, proxyLabels, gateways, listenPort.Port) {
					// Scan all the match blocks
					// if we find any match block without a runtime destination subnet match
					// i.e. match any destination address, then we treat it as the terminal match/catch all match
					// and break out of the loop. We also treat it as a terminal match if the listener is bound
					// to a unix domain socket.
					// But if we find only runtime destination subnet matches in all match blocks, collect them
					// (this is similar to virtual hosts in http) and create filter chain match accordingly.
					if len(match.DestinationSubnets) == 0 || listenPort.Port == 0 {
						out = append(out, &filterChainOpts{
							metadata:         util.BuildConfigInfoMetadata(config.ConfigMeta),
							destinationCIDRs: destinationCIDRs,
							networkFilters:   buildOutboundNetworkFilters(env, node, tcp.Route, push, listenPort, config.ConfigMeta),
						})
						defaultRouteAdded = true
						break TcpLoop
					} else {
						virtualServiceDestinationSubnets = append(virtualServiceDestinationSubnets, match.DestinationSubnets...)
					}
				}
			}

			if len(virtualServiceDestinationSubnets) > 0 {
				out = append(out, &filterChainOpts{
					destinationCIDRs: virtualServiceDestinationSubnets,
					networkFilters:   buildOutboundNetworkFilters(env, node, tcp.Route, push, listenPort, config.ConfigMeta),
				})
			}
		}
	}

	if !defaultRouteAdded {
		// In case of a sidecar config with user defined port, if the user specified port is not the same as the
		// service's port, then pick the service port if and only if the service has only one port. If service
		// has multiple ports, then route to a cluster with the listener port (i.e. sidecar defined port) - the
		// traffic will most likely blackhole.
		port := listenPort.Port
		if len(service.Ports) == 1 {
			port = service.Ports[0].Port
		}

		clusterName := model.BuildSubsetKey(model.TrafficDirectionOutbound, "", service.Hostname, port)
		out = append(out, &filterChainOpts{
			destinationCIDRs: []string{destinationCIDR},
			networkFilters:   buildOutboundNetworkFiltersWithSingleDestination(env, node, clusterName, listenPort),
		})
	}

	return out
}

// This function can be called for namespaces with the auto generated sidecar, i.e. once per service and per port.
// OR, it could be called in the context of an egress listener with specific TCP port on a sidecar config.
// In the latter case, there is no service associated with this listen port. So we have to account for this
// missing service throughout this file
func buildSidecarOutboundTCPTLSFilterChainOpts(env *model.Environment, node *model.Proxy, push *model.PushContext,
	configs []model.Config, destinationCIDR string, service *model.Service, listenPort *model.Port,
	proxyLabels model.LabelsCollection, gateways map[string]bool) []*filterChainOpts {

	out := make([]*filterChainOpts, 0)
	var svcConfigs []model.Config
	if service != nil {
		svcConfigs = getConfigsForHost(service.Hostname, configs)
	} else {
		svcConfigs = configs
	}

	out = append(out, buildSidecarOutboundTLSFilterChainOpts(env, node, push, destinationCIDR, service, listenPort,
		proxyLabels, gateways, svcConfigs)...)
	out = append(out, buildSidecarOutboundTCPFilterChainOpts(env, node, push, destinationCIDR, service, listenPort,
		proxyLabels, gateways, svcConfigs)...)
	return out
}

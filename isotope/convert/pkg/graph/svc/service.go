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

package svc

import (
	"istio.io/tools/isotope/convert/pkg/graph/pct"
	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/size"
	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

// Service describes a service in the service graph.
type Service struct {
	// Name is the DNS-addressable name of the service.
	Name string `json:"name"`

	// Type describes what protocol the service supports (e.g. HTTP, gRPC).
	Type svctype.ServiceType `json:"type,omitempty"`

	// NumReplicas is the number of replicas backing this service.
	NumReplicas int32 `json:"numReplicas,omitempty"`

	// IsEntrypoint indicates that this service is an entrypoint into the service
	// graph, representing a public service.
	IsEntrypoint bool `json:"isEntrypoint,omitempty"`

	// ErrorRate is the percentage chance between 0 and 1 that this service
	// should respond with a 500 server error rather than 200 OK.
	ErrorRate pct.Percentage `json:"errorRate,omitempty"`

	// ResponseSize is the number of bytes in the response body.
	ResponseSize size.ByteSize `json:"responseSize,omitempty"`

	// Script is sequentially called each time the service is called.
	Script script.Script `json:"script,omitempty"`

	// NumRbacPolicies is the number of policies generated for each service.
	NumRbacPolicies int32 `json:"numRbacPolicies"`
}

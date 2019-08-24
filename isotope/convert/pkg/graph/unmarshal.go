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

package graph

import (
	"encoding/json"
	"sync"

	"istio.io/tools/isotope/convert/pkg/graph/pct"
	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/size"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

// UnmarshalJSON converts b into a valid ServiceGraph. See validate() for the
// details on what it means to be "valid".
func (g *ServiceGraph) UnmarshalJSON(b []byte) (err error) {
	metadata := serviceGraphJSONMetadata{Defaults: defaultDefaults}
	err = json.Unmarshal(b, &metadata)
	if err != nil {
		return
	}

	*g, err = parseJSONServiceGraphWithDefaults(b, metadata.Defaults)
	if err != nil {
		return
	}

	err = validate(*g)
	if err != nil {
		return
	}

	return
}

func parseJSONServiceGraphWithDefaults(
	b []byte, defaults defaults) (sg ServiceGraph, err error) {
	withGlobalDefaults(defaults, func() {
		var unmarshallable unmarshallableServiceGraph
		innerErr := json.Unmarshal(b, &unmarshallable)
		if innerErr == nil {
			sg = ServiceGraph(unmarshallable)
		} else {
			err = innerErr
		}
	})
	return
}

// defaultDefaults is a stuttery but validly semantic name for the default
// values when parsing JSON defaults.
var (
	defaultDefaults = defaults{
		Type:        svctype.ServiceHTTP,
		NumReplicas: 1,
	}
	defaultMutex sync.Mutex
)

type serviceGraphJSONMetadata struct {
	Defaults defaults `json:"defaults"`
}

type defaults struct {
	Type            svctype.ServiceType `json:"type"`
	ErrorRate       pct.Percentage      `json:"errorRate"`
	ResponseSize    size.ByteSize       `json:"responseSize"`
	Script          script.Script       `json:"script"`
	RequestSize     size.ByteSize       `json:"requestSize"`
	NumReplicas     int32               `json:"numReplicas"`
	NumRbacPolicies int32               `json:"numRbacPolicies"`
}

func withGlobalDefaults(defaults defaults, f func()) {
	defaultMutex.Lock()

	origDefaultService := svc.DefaultService
	svc.DefaultService = svc.Service{
		Type:            defaults.Type,
		NumReplicas:     defaults.NumReplicas,
		ErrorRate:       defaults.ErrorRate,
		ResponseSize:    defaults.ResponseSize,
		Script:          defaults.Script,
		NumRbacPolicies: defaults.NumRbacPolicies,
	}

	origDefaultRequestCommand := script.DefaultRequestCommand
	for _, req := range script.DefaultRequestCommand.Services {
		req.Size = defaults.RequestSize
	}

	f()

	svc.DefaultService = origDefaultService
	script.DefaultRequestCommand = origDefaultRequestCommand

	defaultMutex.Unlock()
}

type unmarshallableServiceGraph ServiceGraph

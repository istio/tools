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
	"reflect"
	"testing"
	"time"

	"github.com/jmcvetta/randutil"

	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/svc"

	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

func TestServiceGraph_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input []byte
		graph ServiceGraph
		err   error
	}{
		{jsonWithOneService, graphWithOneService, nil},
		{jsonWithDefaultsAndManyServices, graphWithDefaultsAndManyServices, nil},
		{
			jsonWithRequestToUndefinedService,
			ServiceGraph{},
			ErrRequestToUndefinedService{"b"},
		},
		{
			jsonWithNestedConcurrentCommand,
			ServiceGraph{},
			ErrNestedConcurrentCommand,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var graph ServiceGraph
			err := json.Unmarshal(test.input, &graph)
			if err == nil {
				if !reflect.DeepEqual(test.graph, graph) {
					t.Errorf("expected %v; actual %v", test.graph, graph)
				}
			} else {
				if test.err != err {
					t.Errorf("expected %v; actual %v", test.err, err)
				}
			}
		})
	}
}

var (
	jsonWithOneService = []byte(`
		{
			"services": [{"name": "a"}]
		}
	`)
	graphWithOneService = ServiceGraph{[]svc.Service{
		{
			Name:        "a",
			Type:        svctype.ServiceHTTP,
			NumReplicas: 1,
		},
	}}
	jsonWithDefaultsAndManyServices = []byte(`
		{
			"defaults": {
				"errorRate": 0.1,
				"numReplicas": 2,
				"requestSize": 516,
				"responseSize": 128,
				"script": [
					{ "sleep": {"100ms": 100} }
				]
			},
			"services": [
				{
					"name": "a",
					"numReplicas": 5
				},
				{
					"name": "b",
					"script": [
						{
							"call": {
								"service": "a",
								"size": "1KiB"
							}
						},
						{ "sleep": {"10ms": 100} }
					]
				},
				{
					"name": "c",
					"type": "grpc",
					"numReplicas": 1,
					"errorRate": "20%",
					"responseSize": "1K",
					"script": [
						[
							{ "call": "a" },
							{ "call": "b" }
						],
						{ "sleep": {"10ms": 100} }
					]
				}
			]
		}
	`)
	graphWithDefaultsAndManyServices = ServiceGraph{[]svc.Service{
		{
			Name:         "a",
			Type:         svctype.ServiceHTTP,
			NumReplicas:  5,
			ErrorRate:    0.1,
			ResponseSize: 128,
			Script: script.Script([]script.Command{
				script.SleepCommand([]randutil.Choice{{100, 100 * time.Millisecond}}),
			}),
		},
		{
			Name:         "b",
			Type:         svctype.ServiceHTTP,
			NumReplicas:  2,
			ErrorRate:    0.1,
			ResponseSize: 128,
			Script: script.Script([]script.Command{
				script.RequestCommand{ServiceName: "a", Size: 1024},
				script.SleepCommand([]randutil.Choice{{100, 10 * time.Millisecond}}),
			}),
		},
		{
			Name:         "c",
			Type:         svctype.ServiceGRPC,
			NumReplicas:  1,
			ErrorRate:    0.2,
			ResponseSize: 1024,
			Script: script.Script([]script.Command{
				script.ConcurrentCommand{
					script.RequestCommand{ServiceName: "a", Size: 516},
					script.RequestCommand{ServiceName: "b", Size: 516},
				},
				script.SleepCommand([]randutil.Choice{{100, 10 * time.Millisecond}}),
			}),
		},
	}}
	jsonWithRequestToUndefinedService = []byte(`
		{
			"services": [
				{
					"name": "a",
					"script": [{ "call": "b"}]
				}
			]
		}
	`)
	jsonWithNestedConcurrentCommand = []byte(`
		{
			"services": [
				{
					"name": "a"
				},
				{
					"name": "b",
					"script": [
						[
							[{ "call": "a" }, { "call": "a" }],
							{ "sleep": {"10ms": 100} }
						]
					]
				}
			]
		}
	`)
)

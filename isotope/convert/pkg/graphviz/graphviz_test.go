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

package graphviz

import (
	"reflect"
	"testing"
	"time"

	"istio.io/tools/isotope/convert/pkg/graph"
	"istio.io/tools/isotope/convert/pkg/graph/msg"
	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/size"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

func TestServiceGraphToGraph(t *testing.T) {
	expected := Graph{
		Nodes: []Node{
			{
				Name:         "a",
				Type:         "HTTP",
				ErrorRate:    "0.01%",
				ResponseSize: "{\"type\":\"static\",\"data\":{\"size\":\"10KiB\",\"number\":1}}",
				Steps: [][]string{
					{
						"SLEEP 100ms",
					},
				},
			},
			{
				Name:         "b",
				Type:         "gRPC",
				ErrorRate:    "0.00%",
				ResponseSize: "{\"type\":\"static\",\"data\":{\"size\":\"10KiB\",\"number\":1}}",
				Steps:        [][]string{},
			},
			{
				Name:         "c",
				Type:         "HTTP",
				ErrorRate:    "0.00%",
				ResponseSize: "{\"type\":\"static\",\"data\":{\"size\":\"10KiB\",\"number\":1}}",
				Steps: [][]string{
					{
						"CALL \"a\" 10KiB",
					},
					{
						"CALL \"b\" 1KiB",
					},
				},
			},
			{
				Name:         "d",
				Type:         "HTTP",
				ErrorRate:    "0.00%",
				ResponseSize: "{\"type\":\"static\",\"data\":{\"size\":\"10KiB\",\"number\":1}}",
				Steps: [][]string{
					{
						"CALL \"a\" 1KiB",
						"CALL \"c\" 1KiB",
					},
					{
						"SLEEP 10ms",
					},
					{
						"CALL \"b\" 1KiB",
					},
				},
			},
		},
		Edges: []Edge{
			{
				From:      "c",
				To:        "a",
				StepIndex: 0,
			},
			{
				From:      "c",
				To:        "b",
				StepIndex: 1,
			},
			{
				From:      "d",
				To:        "a",
				StepIndex: 0,
			},
			{
				From:      "d",
				To:        "c",
				StepIndex: 0,
			},
			{
				From:      "d",
				To:        "b",
				StepIndex: 2,
			},
		},
	}

	serviceGraph := graph.ServiceGraph{
		Services: []svc.Service{
			{
				Name:         "a",
				Type:         svctype.ServiceHTTP,
				ErrorRate:    0.0001,
				ResponseSize: msg.MessageSize{"static", msg.MessageSizeStatic{size.ByteSize(10240), 1}},
				Script: []script.Command{
					script.SleepCommand(100 * time.Millisecond),
				},
			},
			{
				Name:         "b",
				Type:         svctype.ServiceGRPC,
				ErrorRate:    0,
				ResponseSize: msg.MessageSize{"static", msg.MessageSizeStatic{size.ByteSize(10240), 1}},
			},
			{
				Name:         "c",
				Type:         svctype.ServiceHTTP,
				ErrorRate:    0,
				ResponseSize: msg.MessageSize{"static", msg.MessageSizeStatic{size.ByteSize(10240), 1}},
				Script: []script.Command{
					script.RequestCommand{
						ServiceName: "a",
						Size:        10240,
					},
					script.RequestCommand{
						ServiceName: "b",
						Size:        1024,
					},
				},
			},
			{
				Name:         "d",
				Type:         svctype.ServiceHTTP,
				ErrorRate:    0,
				ResponseSize: msg.MessageSize{"static", msg.MessageSizeStatic{size.ByteSize(10240), 1}},
				Script: []script.Command{
					script.ConcurrentCommand([]script.Command{
						script.RequestCommand{
							ServiceName: "a",
							Size:        1024,
						},
						script.RequestCommand{
							ServiceName: "c",
							Size:        1024,
						},
					}),
					script.SleepCommand(10 * time.Millisecond),
					script.RequestCommand{
						ServiceName: "b",
						Size:        1024,
					},
				},
			},
		},
	}
	actual, err := ServiceGraphToGraph(serviceGraph)
	if err != nil {
		t.Fatal(err)
	}
	if !graphsAreEqual(expected, actual) {
		t.Errorf("\nexpect: %+v, \nactual: %+v", expected, actual)
	}
}

func graphsAreEqual(left Graph, right Graph) bool {
	return reflect.DeepEqual(left, right)
}

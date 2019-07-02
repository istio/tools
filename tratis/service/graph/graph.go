// Copyright 2019 Istio Authors
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
	"log"
	"sort"
	"strings"

	jaeger "github.com/jaegertracing/jaeger/model/json"
)

type NodeData struct {
	SpanID        jaeger.SpanID `json:"spanID,omitempty"`
	OperationName string        `json:"operationName,omitempty"`
	StartTime     uint64        `json:"startTime,omitempty"`
	Duration      uint64        `json:"duration,omitempty"`
	RequestType   string        `json:"requestType"`
}

type Node struct {
	Data     NodeData `json:"data"`
	Children *[]Node  `json:"children"`
}

type Graph struct {
	Root *Node `'json:"root"`
}

func (g *Graph) ExtractGraphData() []byte {
	bytes, err := json.Marshal(g.Root)

	if err != nil {
		log.Fatalf(`graph structure is improper`)
	}

	return bytes
}

func findTag(tags []jaeger.KeyValue, key string) jaeger.KeyValue {
	for _, tag := range tags {
		if tag.Key == key {
			return tag
		}
	}

	return jaeger.KeyValue{}
}


// Root span has no references.
func findRootSpan(spans []jaeger.Span) jaeger.Span {
	for _, span := range spans {
		if len(span.References) == 0 {
			return span
		}
	}

	log.Fatalf(`Root Span not present in spans`)
	return jaeger.Span{}
}

func GenerateGraph(data []jaeger.Span) *Graph {
	rootSpan := findRootSpan(data)

	tag := findTag(rootSpan.Tags, "upstream_cluster")
	tagData := strings.Split(tag.Value.(string), "|")[0]

	d := NodeData{rootSpan.SpanID, rootSpan.OperationName,
		rootSpan.StartTime, rootSpan.Duration, tagData}
	childrenList := UpdateChildren(data, rootSpan.SpanID)
	root := Node{d, &childrenList}
	return &Graph{&root}
}

func UpdateChildren(data []jaeger.Span, spanID jaeger.SpanID) []Node {
	children := make([]Node, 0)

	for _, v := range data {
		if len(v.References) == 0 {
			continue
		}

		ref := v.References[0]
		if ref.RefType == jaeger.ChildOf && ref.SpanID == spanID {
			/*
			Tag Examples:

			{Key: upstream_cluster 
			 Type: string 
			 Value: inbound|9080|http|productpage.default.svc.cluster.local
			}

			{Key: upstream_cluster 
			 Type: string 
			 Value: inbound|9080|http|reviews.default.svc.cluster.local
			} 

			{Key: upstream_cluster
			 Type: string
			 Value: outbound|9080||ratings.default.svc.cluster.local
			}
			*/

			tag := findTag(v.Tags, "upstream_cluster")
			tagData := strings.Split(tag.Value.(string), "|")[0]

			d := NodeData{v.SpanID, v.OperationName,
				v.StartTime, v.Duration, tagData}

			nodeChildren := UpdateChildren(data, v.SpanID)
			children = append(children, Node{d, &nodeChildren})
		}
	}


	sort.Slice(children, func(i, j int) bool {
		return (children[i].Data.StartTime <
			children[j].Data.StartTime)
	})

	return children
}

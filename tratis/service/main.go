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

package main

import (
	"fmt"
	"log"
	"os"

	"istio.io/tools/tratis/service/distribution"
	"istio.io/tools/tratis/service/graph"
	parser "istio.io/tools/tratis/service/parsing"
	"istio.io/tools/tratis/service/pkg/consts"
)

func addNewGraph(data *[][]*graph.Graph, g *graph.Graph) int {
	ret := -1
	for idx, graphs := range *data {
		if graph.CompGraph(graphs[0], g) {
			ret = idx
			break
		}
	}

	if ret == -1 {
		temp := make([]*graph.Graph, 0)
		temp = append(temp, g)
		*data = append(*data, temp)

		return (len(*data) - 1)
	}
	(*data)[ret] = append((*data)[ret], g)
	return ret
}

func main() {
	fmt.Println("Starting Tratis ...")

	if len(os.Args) != 2 {
		log.Fatalf(`Input Arguments not correctly provided: go run main.go <TOOL_NAME=jaeger/zipkin>`)
	}

	TracingToolName := os.Args[1]

	fmt.Println("Generating Traces ...")

	data, err := parser.ParseJSON(TracingToolName)
	if err != nil {
		log.Fatalf(`Connection between "%s" and tratis is broken`,
			TracingToolName)
	}

	fmt.Println("Filtering Traces ...")

	traces := data.Traces

	fmt.Printf("Processing %d Traces\n", len(traces))

	d := make([][][]distribution.TimeInformation, 0)
	graphs := make([][]*graph.Graph, 0)

	fmt.Println("Generating Time Information ...")

	for _, trace := range traces {
		g := graph.GenerateGraph(trace.Spans)
		idx := addNewGraph(&graphs, g)

		traceInformation := distribution.ExtractTimeInformation(g)

		if len(d) < idx+1 {
			d = append(d, make([][]distribution.TimeInformation, 0))
		}

		d[idx] = append(d[idx], traceInformation)
	}

	fmt.Println("Combining Results + Distribution Fitting ...")

	for idx := range graphs {
		fmt.Println("=======================================================")
		fmt.Println("Number of Traces: ", len(graphs[idx]))
		fmt.Println("Call Graph: ", string(graphs[idx][0].ExtractGraphData()))

		combinedResults := distribution.CombineTimeInformation(d[idx])
		dists := distribution.TimeInfoToDist(consts.DistFilePath,
			consts.DistFittingFuncName, combinedResults)
		fmt.Println("Distribution Details: ", dists)
		fmt.Println("=======================================================")
	}
}

/*
	LOGIC:
		FOR EACH GRAPH CHECK LISTS OF LISTS.
		IF GRAPH DOES NOT MATCH MAKE NEW CATEGORY <==> OTHERWISE PUSH INTO OLDER CATEGORY
		USE SAME INDEX FOR TIME INFO
		HENCE MULTIPLE CALL GRAPHS WITH DIFFERENT TIME INFO.
*/

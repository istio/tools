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

	jaeger "github.com/jaegertracing/jaeger/model/json"
	"istio.io/tools/tratis/service/distribution"
	"istio.io/tools/tratis/service/graph"
	parser "istio.io/tools/tratis/service/parsing"
	"istio.io/tools/tratis/service/pkg/consts"
)

func filterTracesByNumSpans(traces []jaeger.Trace) []jaeger.Trace {
	ret := make([]jaeger.Trace, 0)

	for _, trace := range traces {
		if len(trace.Spans) == consts.NumberSpans {
			ret = append(ret, trace)
		}
	}

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
	traces = filterTracesByNumSpans(traces)

	fmt.Printf("Processing %d Traces\n", len(traces))

	d := make([][]distribution.TimeInformation, 0)

	fmt.Println("Generating Time Information ...")

	for idx, trace := range traces {
		g := graph.GenerateGraph(trace.Spans)

		if idx == 0 {

			fmt.Println("Call Graph: ", string(g.ExtractGraphData()))
		}

		traceInformation := distribution.ExtractTimeInformation(g)

		d = append(d, traceInformation)
	}

	fmt.Println("Combining Results + Distribution Fitting ...")

	combinedResults := distribution.CombineTimeInformation(d)
	dists := distribution.TimeInfoToDist(consts.DistFilePath,
		consts.DistFittingFuncName, combinedResults)
	fmt.Println("Distribution Details: ", dists)
}

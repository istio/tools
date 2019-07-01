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

package main

import (
	"fmt"
	"log"
	"os"
	// "path"

	jaeger "github.com/jaegertracing/jaeger/model/json"
	"istio.io/tools/tratis/service/distribution"
	"istio.io/tools/tratis/service/graph"
	parser "istio.io/tools/tratis/service/parsing"
	"istio.io/tools/tratis/service/pkg/consts"
)

func cleanTraces(traces []jaeger.Trace) []jaeger.Trace {
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

	TracingToolName, ok := os.LookupEnv(consts.TracingToolEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.TracingToolEnvKey)
	}

	data, err := parser.ParseJSON(TracingToolName)
	if err != nil {
		log.Fatalf(`Connection between "%s" and tratis is broken`,
			TracingToolName)
	}

	traces := data.Traces
	traces = cleanTraces(traces)

	d := make([][]distribution.TimeInformation, 0)

	for _, trace := range traces {
		g := graph.GenerateGraph(trace.Spans)
		traceInformation := distribution.ExtractTimeInformation(g)

		d = append(d, traceInformation)
	}

	combinedResults := distribution.CombineTimeInformation(d)
	distribution.TimeInfoToDist("Distribution", "BestFitDistribution", combinedResults)
}

//

/*

TODO LIST:

1. RUN GRAPH FILE
2. FIX DISTRIBUTION FILE

*/

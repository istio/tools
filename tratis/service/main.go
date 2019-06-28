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
	"io/ioutil"
	"log"
	"os"
	"path"

	// "istio.io/tools/tratis/service/distribution"
	"istio.io/tools/tratis/service/graph"

	parser "istio.io/tools/tratis/service/parsing"
	"istio.io/tools/tratis/service/pkg/consts"
)

func main() {
	TracingToolName, ok := os.LookupEnv(consts.TracingToolEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.TracingToolEnvKey)
	}

	traces, err := ioutil.ReadDir(consts.TraceFilesPath)

	if err != nil {
		log.Fatal(err)
	}

	for _, t := range traces {
		traceFilePath := path.Join(consts.TraceFilesPath, t.Name())
		trace, err := parser.ParseJSON(traceFilePath, TracingToolName)

		if err != nil {
			log.Fatalf(`trace file "%s" is not correctly formatted`,
				traceFilePath)
		}

		fmt.Println(trace)
		fmt.Println(".. ..")

		g := graph.GenerateGraph(trace.Spans)
		// fmt.Println(string(distribution.ExtractTimeInformation(g)))
	}
}

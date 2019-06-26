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
	"path"

	"istio.io/tools/tratis/service/graph"

	parser "istio.io/tools/tratis/service/parsing"
	"istio.io/tools/tratis/service/pkg/consts"
)

var (
	ApplicationTraceJSONFilePath = path.Join(
		consts.ConfigPath, consts.ApplicationTraceJSONFileName)
)

func main() {
	TracingToolName, ok := os.LookupEnv(consts.TracingToolEnvKey)
	if !ok {
		log.Fatalf(`env var "%s" is not set`, consts.TracingToolEnvKey)
	}

	trace, err := parser.ParseJSON(ApplicationTraceJSONFilePath,
		TracingToolName)

	if err != nil {
		log.Fatalf(`trace file "%s" is not correctly formatted`,
			ApplicationTraceJSONFilePath)
	}

	g := graph.GenerateGraph(trace.Spans)
	fmt.Println(string(g.ExtractGraphData()))
}

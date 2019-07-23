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

	"istio.io/tools/tratis/service/output"
	parser "istio.io/tools/tratis/service/parsing"
)

func main() {
	fmt.Println("Starting Tratis ...")

	if len(os.Args) != 3 {
		log.Fatalf(`Input Arguments not correctly provided: go run main.go <TOOL_NAME=jaeger/zipkin> <RESULTS_JSON_FILE>`)
	}

	TracingToolName := os.Args[1]
	jsonFileName := os.Args[2]

	fmt.Println("Generating Traces ...")

	data, err := parser.ParseJSON(TracingToolName)
	if err != nil {
		log.Fatalf(`Connection between "%s" and tratis is broken`,
			TracingToolName)
	}

	results := output.GenerateOutput(data)

	f, err := os.Create(jsonFileName)
	if err != nil {
		log.Fatalf("Unable to create %s: %v", jsonFileName, err)
	}
	f.Write(append(results, '\n'))
}

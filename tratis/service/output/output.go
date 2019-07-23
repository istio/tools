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

package output

import (
	"encoding/json"

	"istio.io/tools/tratis/service/distribution"
	"istio.io/tools/tratis/service/graph"
	parser "istio.io/tools/tratis/service/parsing"
	"istio.io/tools/tratis/service/pkg/consts"

	"fmt"
)

type Output struct {
	NumTraces               int                               `'json"NumTraces"`
	CallGraph               *graph.Node                       `'json:"Graph"`
	TimeInformation         []distribution.TotalDistributions `json:"TimeInformation"`
	RequestSizeInformation  []distribution.TotalDistributions `json:"RequestSizeInformation"`
	ResponseSizeInformation []distribution.TotalDistributions `json:"ResponseSizeInformation"`
}

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

func GenerateOutput(data parser.TraceData) []byte {
	traces := data.Traces

	fmt.Printf("Processing %d Traces\n", len(traces))

	d := make([][][]distribution.TimeInformation, 0)
	s := make([][][]distribution.MessageSizeInfo, 0)

	graphs := make([][]*graph.Graph, 0)

	fmt.Println("Generating Time Information ...")

	for _, trace := range traces {
		g := graph.GenerateGraph(trace.Spans)
		idx := addNewGraph(&graphs, g)

		traceInformation := distribution.ExtractTimeInformation(g)
		sizeInformation := distribution.ExtractSizeInformation(g)

		if len(d) < idx+1 {
			d = append(d, make([][]distribution.TimeInformation, 0))
		}

		if len(s) < idx+1 {
			s = append(s, make([][]distribution.MessageSizeInfo, 0))
		}

		d[idx] = append(d[idx], traceInformation)
		s[idx] = append(s[idx], sizeInformation)
	}

	fmt.Println("Combining Results + Distribution Fitting ...")

	ret := make([]Output, 0)

	for idx := range graphs {

		if len(graphs[idx]) > consts.MinNumTraces {
			fmt.Println("Processing Graph Number: ", idx)
			temp := Output{}
			temp.NumTraces = len(graphs[idx])
			temp.CallGraph = graphs[idx][0].Root

			fmt.Println("Extra Time Information")

			timeResults := distribution.ConvertTimeInfo(distribution.CombineTimeInformation(d[idx]))

			dists := distribution.InfoToDist(consts.DistFilePath,
				consts.DistFittingFuncName, timeResults)

			temp.TimeInformation = dists

			fmt.Println("Extracting Response Size Information")

			sizeResults := distribution.ConvertSizeInfo(distribution.CombineSizeInformation(s[idx], true))

			dists = distribution.InfoToDist(consts.DistFilePath,
				consts.DistFittingFuncName, sizeResults)

			temp.ResponseSizeInformation = dists

			fmt.Println("Extracting Request Size Information")

			sizeResults = distribution.ConvertSizeInfo(distribution.CombineSizeInformation(s[idx], false))

			dists = distribution.InfoToDist(consts.DistFilePath,
				consts.DistFittingFuncName, sizeResults)

			temp.RequestSizeInformation = dists

			// Appending to ret

			ret = append(ret, temp)
		} else {
			fmt.Print("Skipping Graph Number: ", idx)
		}
	}

	bytes, _ := json.MarshalIndent(ret, "", "  ")
	return bytes
}

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

package distribution

import (
	"encoding/json"
	// jaeger "github.com/jaegertracing/jaeger/model/json"
	"istio.io/tools/tratis/service/graph"
)

type Time struct {
	StartTime uint64 `json:"startTime"`
	EndTime   uint64 `json:"endTime"`
	Duration  uint64 `json:"duration`
}

type TimeInformation struct {
	TimeData      []Time `json:"time"`
	OperationName string `json:"operationName,omitempty"`
}

func ExtractTimeInformation(g *graph.Graph) []byte {
	ret := make([]TimeInformation, 0)
	ExtractTimeInformationWrapper(g.Root, &ret)

	bytes, _ := json.Marshal(ret)
	return bytes
}

func ExtractTimeInformationWrapper(n *graph.Node, t *[]TimeInformation) {
	if n == nil {
		return
	}

	nodeStartTime := n.Data.StartTime
	nodeEndTime := n.Data.StartTime + n.Data.Duration

	timeData := make([]Time, 0)

	for _, child := range *n.Children {
		d := child.Data.StartTime - nodeStartTime
		newTime := Time{nodeStartTime, child.Data.StartTime, d}
		timeData = append(timeData, newTime)
		nodeStartTime = child.Data.StartTime + child.Data.Duration

		ExtractTimeInformationWrapper(&child, t)
	}

	d := nodeEndTime - nodeStartTime
	newTime := Time{nodeStartTime, nodeEndTime, d}
	timeData = append(timeData, newTime)

	if n.Data.RequestType == "inbound" {
		*t = append(*t, TimeInformation{timeData, n.Data.OperationName})
	}
}

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
	"istio.io/tools/tratis/service/graph"
)

type Time struct {
	StartTime int `json:"startTime"`
	EndTime   int `json:"endTime"`
}

type TimeInformation struct {
	TimeData []Time         `json:"time"`
	Data     graph.NodeData `json:"data"`
}

func ExtractTimeInformation(g *graph.Graph) []TimeInformation {
	ret := make([]TimeInformation, 0)
	ExtractTimeInformationWrapper(g.Root, &ret)
	return ret
}

func ExtractTimeInformationWrapper(n *graph.Node, t *[]TimeInformation) {
	if n == nil {
		return
	}

	nodeStartTime := n.Data.StartTime
	nodeEndTime := n.Data.StartTime + n.Data.Duration

	timeData := make([]Time, 0)

	for _, child := range *n.Children {
		newTime := Time{nodeStartTime, child.Data.StartTime}
		timeData = append(timeData, newTime)
		nodeStartTime = child.Data.StartTime + child.Data.Duration

		ExtractTimeInformationWrapper(&child, t)
	}

	newTime := Time{nodeStartTime, nodeEndTime}
	timeData = append(timeData, newTime)

	if n.Data.ReqType == "inbound" {
		*t = append(*t, TimeInformation{timeData, n.Data})
	}
}

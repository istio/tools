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
	// "fmt"
	// "encoding/json"
	// jaeger "github.com/jaegertracing/jaeger/model/json"
	"istio.io/tools/tratis/service/graph"
	"istio.io/tools/tratis/service/pkg/consts"
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

type CombinedTimeInformation struct {
	Duration      [][]uint64 `json:"durations"`
	OperationName string     `json:"operationName"`
}

func CombineTimeInformation(data [][]TimeInformation) []CombinedTimeInformation {
	ret := make([]CombinedTimeInformation, consts.NumberServices)

	for _, trace := range data {
		for idx, span := range trace {
			ret[idx].OperationName = span.OperationName

			if len(ret[idx].Duration) == 0 {
				ret[idx].Duration = make([][]uint64, len(span.TimeData))
			}

			for timeIndex, duration := range span.TimeData {
				ret[idx].Duration[timeIndex] =
					append(ret[idx].Duration[timeIndex], duration.Duration)
			}
		}
	}

	return ret
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

// func CombineTimeInformation(data [][]TimeInformation) []CombineTimeInformation {
// 	// [{[{1562002522923573 1562002522925988 2415}] details.default.svc.cluster.local:9080/*} {[{1562002522933506 1562002522934836 1330}] ratings.default.svc.cluster.local:9080/*} {[{1562002522930153 1562002522933052 2899} {1562002522935123 1562002522936216 1093}] reviews.default.svc.cluster.local:9080/*} {[{1562002522919404 1562002522923162 3758} {1562002522926485 1562002522929816 3331} {1562002522937008 1562002522939600 2592}] productpage.default.svc.cluster.local:9080/productpage}]

// 	// 1. Add Operation Name
// 	// 2. Add Time Datax`

// 	ret := make([]CombineTimeInformation, consts.NumberSpans)

// 	for _, ret := range ret {
// 		ret.Duration
// 	}

// 	for _, timeInfo := range data {
// 		for idx, span := range timeInfo {
// 			ret[idx].OperationName = span.OperationName
// 			ret[idx]

// 			for idx_timeData,
// 		}
// 	}
// }

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

package span

import (
	"encoding/json"
	"fmt"

	"istio.io/tools/tratis/service/parsing/pkg/reference"
	"istio.io/tools/tratis/service/parsing/pkg/tag"
)

// UnmarshalJSON converts b to a Service, applying the default values from
// DefaultService.
func (span *Span) UnmarshalJSON(b []byte) (err error) {
	var spanData unmarshableSpan
	err = json.Unmarshal(b, &spanData)
	if err != nil {
		fmt.Println(err)
		return
	}

	span.TraceID = spanData.TraceID
	span.SpanID = spanData.SpanID
	span.OperationName = spanData.OperationName
	span.StartTime = spanData.StartTime
	span.Duration = spanData.Duration

	span.Tags = make(tag.Tag)

	for _, t := range spanData.Tags {
		span.Tags[t["key"].(string)] = t["value"].(string)
	}

	span.References = spanData.References
	span.ProcessID = spanData.ProcessID
	span.Logs = spanData.Logs
	span.Warnings = spanData.Warnings

	return
}

type unmarshableSpan struct {
	TraceID       string                `json:"traceID"`
	SpanID        string                `json:"spanID"`
	OperationName string                `json:"operationName"`
	StartTime     int                   `json:"startTime"`
	Duration      int                   `json:"duration"`
	Tags          []tag.Tag             `json:"tags"`
	References    []reference.Reference `json:"references"`
	ProcessID     string                `json:"processID"`
	Logs          []string              `json:"logs"`
	Warnings      []string              `json:"warnings"`
}

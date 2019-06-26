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

package trace

import (
	"encoding/json"
	"fmt"

	"istio.io/tools/tratis/service/parsing/pkg/process"
	"istio.io/tools/tratis/service/parsing/pkg/span"
)

// UnmarshalJSON converts b to a Service, applying the default values from
// DefaultService.
func (trace *Trace) UnmarshalJSON(b []byte) (err error) {
	var traceData unmarshallableTrace
	err = json.Unmarshal(b, &traceData)
	if err != nil {
		fmt.Println(err)
		return
	}

	trace.TraceID = traceData.TraceID
	trace.Spans = make(map[string]span.Span)

	for _, s := range traceData.Spans {
		trace.Spans[s.SpanID] = s
	}

	trace.Processes = traceData.Processes

	return
}

type unmarshallableTrace struct {
	TraceID   string                     `json:"traceID"`
	Spans     []span.Span                `json:"spans"`
	Processes map[string]process.Process `'json:"processes"`
}

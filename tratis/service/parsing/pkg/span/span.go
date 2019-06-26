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
	"istio.io/tools/tratis/service/parsing/pkg/reference"
	"istio.io/tools/tratis/service/parsing/pkg/tag"
)

// Span describes a single span (distributed-tracing term) and
// its associated data.
type Span struct {
	TraceID       string                `json:"traceID"`
	SpanID        string                `json:"spanID"`
	OperationName string                `json:"operationName"`
	StartTime     int                   `json:"startTime"`
	Duration      int                   `json:"duration"`
	Tags          tag.Tag               `json:"tags"`
	References    []reference.Reference `json:"references"`
	ProcessID     string                `json:"processID"`
	Logs          []string              `json:"logs"`
	Warnings      []string              `json:"warnings"`
}

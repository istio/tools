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

package consts

const (
	// TracingToolEnvKey is the key of the environment variable whose value is
	// the name of the service.
	TracingToolEnvKey = "TOOL_NAME"

	// An application might generate traces with varying number of
	// spans. Only the traces with NumberSpans would be picked.
	NumberSpans = 8

	// Max Number of Services in Application
	NumberServices = 4

	// Jaeger URL
	JaegerURL = "http://localhost:15034"
	// Service Name
	ServiceName = ""
	// Traces Limit
	NumTraces = 1000

	// Distribution Fitting File Path
	DistFilePath = "Distribution"

	// Distribution Fitting Function Name
	DistFittingFuncName = "BestFitDistribution"

	// Tracing Tool IP Address
	TracingToolAddress = "localhost"
	// Tarcing Tool Port Number
	TracingToolPortNumber = "15034"
	// Tracing Tool EntryPoint Service
	TracingToolEntryPoint = "istio-ingressgateway"
)

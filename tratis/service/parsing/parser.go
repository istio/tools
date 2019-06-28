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

package parser

import (
	"encoding/json"
	"io/ioutil"
	"log"

	jaeger "github.com/jaegertracing/jaeger/model/json"
)

type traceData struct {
	Traces []jaeger.Trace `json:"data"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func ParseJSON(filePath string, toolName string) (appTrace jaeger.Trace,
	err error) {
	traceJSON, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	if toolName == "jaeger" {
		return parseJaeger(traceJSON)
	} else if toolName == "zipkin" {
		return parseZipkin(traceJSON)
	}

	log.Fatalf(`tracing tool "%s" is not correctly supported`, toolName)
	return
}

func parseJaeger(data []byte) (appTrace jaeger.Trace, err error) {
	t := traceData{}
	err = json.Unmarshal(data, &t)
	if err != nil {
		return
	}

	return t.Traces[0], nil
}

func parseZipkin(data []byte) (appTrace jaeger.Trace, err error) {
	err = json.Unmarshal(data, &appTrace)
	if err != nil {
		return
	}
	return appTrace, nil
}

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

package graph

import (
	"reflect"
	"testing"

	jaeger "github.com/jaegertracing/jaeger/model/json"
)

var (
	tags = []jaeger.KeyValue{{Key: "component", Type: "string", Value: "proxy"},
		{Key: "node_id", Type: "string", Value: "sidecar~10.20.0.5~productpage-v1-6c668694dc-dmnsl.default~default.svc.cluster.local"},
		{Key: "zone", Type: "string", Value: "us-central1-a"},
		{Key: "guid:x-request-id", Type: "string", Value: "1f444b8f-9dc1-9afb-b5b3-fba527800194"},
		{Key: "http.url", Type: "string", Value: "http://details:9080/details/0"},
		{Key: "http.method", Type: "string", Value: "GET"},
		{Key: "downstream_cluster", Type: "string", Value: "-"},
		{Key: "user_agent", Type: "string", Value: "curl/7.64.0"},
		{Key: "http.protocol", Type: "string", Value: "HTTP/1.1"},
		{Key: "request_size", Type: "string", Value: "0"},
		{Key: "upstream_cluster", Type: "string", Value: "outbound|9080||details.default.svc.cluster.local"},
		{Key: "http.status_code", Type: "string", Value: "200"},
		{Key: "response_size", Type: "string", Value: "178"},
		{Key: "response_flags", Type: "string", Value: "-"},
	}
)

func TestfindTag(t *testing.T) {

	var tests = []struct {
		actual   jaeger.KeyValue
		expected jaeger.KeyValue
		msg      string
	}{
		{jaeger.KeyValue{Key: "response_size", Type: "string", Value: "178"},
			findTag(tags, "response_size"),
			""},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			if reflect.DeepEqual(test.actual, test.expected) {
				t.Errorf("%s: got %+v, not as expected %+v", test.msg, test.actual, test.expected)
			}
		})
	}
}

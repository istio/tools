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

package svc

import (
	"encoding/json"
	"testing"

	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

func TestService_MarshalJSON(t *testing.T) {
	tests := []struct {
		input  Service
		output []byte
		err    error
	}{
		{
			Service{
				Name:        "a",
				Type:        svctype.ServiceHTTP,
				NumReplicas: 1,
			},
			[]byte(`{"name":"a","type":"http","numReplicas":1,"responseSize":{"type":"","data":null},"numRbacPolicies":0}`),
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			output, err := json.Marshal(test.input)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if string(test.output) != string(output) {
				t.Errorf("expected %s; actual %s", test.output, output)
			}
		})
	}
}

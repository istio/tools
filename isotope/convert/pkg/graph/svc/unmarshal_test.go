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
	"reflect"
	"testing"

	"istio.io/tools/isotope/convert/pkg/graph/svctype"
)

func TestService_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input []byte
		svc   Service
		err   error
	}{
		{
			[]byte(`{"name": "A"}`),
			Service{
				Name:        "A",
				Type:        svctype.ServiceHTTP,
				NumReplicas: 1,
			},
			nil,
		},
		{
			[]byte(`{}`),
			Service{Type: svctype.ServiceHTTP, NumReplicas: 1},
			ErrEmptyName,
		},
		{
			[]byte(`{"name": ""}`),
			Service{Type: svctype.ServiceHTTP, NumReplicas: 1},
			ErrEmptyName,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var svc Service
			err := json.Unmarshal(test.input, &svc)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if !reflect.DeepEqual(test.svc, svc) {
				t.Errorf("expected %v; actual %v", test.svc, svc)
			}
		})
	}
}

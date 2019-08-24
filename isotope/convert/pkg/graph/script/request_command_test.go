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

package script

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestRequestCommand_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input   []byte
		command RequestCommand
		err     error
	}{
		{
			[]byte(`{"Services": [{"service": "A", "probability": 100}]}`),
			RequestCommand{Services: []RequestCommandData{{ServiceName: "A", Probability: 100}}},
			nil,
		},
		{
			[]byte(`{"Services": [{"service": "A", "probability": 100, "size": 128}]}`),
			RequestCommand{Services: []RequestCommandData{{ServiceName: "A", Probability: 100, Size: 128}}},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var command RequestCommand
			err := json.Unmarshal(test.input, &command)

			fmt.Println(test.input)
			fmt.Println(&command)

			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if !reflect.DeepEqual(test.command, command) {
				t.Errorf("expected %v; actual %v", test.command, command)
			}
		})
	}
}

func TestRequestCommand_UnmarshalJSON_Default(t *testing.T) {
	tests := []struct {
		input   []byte
		command RequestCommand
		err     error
	}{
		{
			[]byte(`{"Services": [{"service": "A", "probability": 100, "size": 512}]}`),
			RequestCommand{Services: []RequestCommandData{{ServiceName: "A", Probability: 100, Size: 512}}},
			nil,
		},
		{
			[]byte(`{"Services": [{"service": "A", "probability": 100, "size": 128}]}`),
			RequestCommand{Services: []RequestCommandData{{ServiceName: "A", Probability: 100, Size: 128}}},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			// t.Parallel()

			var command RequestCommand
			err := json.Unmarshal(test.input, &command)

			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if !reflect.DeepEqual(test.command, command) {
				t.Errorf("expected %v; actual %v", test.command, command)
			}
		})
	}
}

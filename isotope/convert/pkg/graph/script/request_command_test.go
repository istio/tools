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
	"testing"
)

func TestRequestCommand_UnmarshalJSON(t *testing.T) {
	DefaultRequestCommand = RequestCommand{}

	tests := []struct {
		input   []byte
		command RequestCommand
		err     error
	}{
		{
			[]byte(`"A"`),
			RequestCommand{ServiceName: "A"},
			nil,
		},
		{
			[]byte(`{"service": "A"}`),
			RequestCommand{ServiceName: "A"},
			nil,
		},
		{
			[]byte(`{"service": "a", "size": 128}`),
			RequestCommand{ServiceName: "a", Size: 128},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var command RequestCommand
			err := json.Unmarshal(test.input, &command)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if test.command != command {
				t.Errorf("expected %v; actual %v", test.command, command)
			}
		})
	}
}

func TestRequestCommand_UnmarshalJSON_Default(t *testing.T) {
	DefaultRequestCommand = RequestCommand{Size: 512}

	tests := []struct {
		input   []byte
		command RequestCommand
		err     error
	}{
		{
			[]byte(`"A"`),
			RequestCommand{ServiceName: "A", Size: 512},
			nil,
		},
		{
			[]byte(`{"service": "A"}`),
			RequestCommand{ServiceName: "A", Size: 512},
			nil,
		},
		{
			[]byte(`{"service": "a", "size": 128}`),
			RequestCommand{ServiceName: "a", Size: 128},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var command RequestCommand
			err := json.Unmarshal(test.input, &command)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if test.command != command {
				t.Errorf("expected %v; actual %v", test.command, command)
			}
		})
	}
}

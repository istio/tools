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
	"reflect"
	"testing"
	"time"
)

func TestScript_UnmarshalJSON(t *testing.T) {
	DefaultRequestCommand = RequestCommand{}

	tests := []struct {
		input  []byte
		script Script
		err    error
	}{
		{
			[]byte(`[]`),
			Script{},
			nil,
		},
		{
			[]byte(`[{"sleep": {"SleepCommand": [{"Load": {"Min": 0, "Max": 100}, "type": "static", "data": {"time": "1s"}}]}}]`),
			Script{
				SleepCommand{[]SleepCommandData{SleepCommandData{Range{uint64(0),uint64(100)}, Static, SleepCommandStatic{1 * time.Second}}}},
			},
			nil,
		},
		{
			[]byte(`[{"call": "A"}, {"sleep": {"SleepCommand": [{"Load": {"Min": 0, "Max": 100}, "type": "static", "data": {"time": "10ms"}}]}}]`),
			Script{
				RequestCommand{ServiceName: "A"},
				SleepCommand{[]SleepCommandData{SleepCommandData{Range{uint64(0),uint64(100)}, Static, SleepCommandStatic{10 * time.Millisecond}}}},
			},
			nil,
		},
		{
			[]byte(`[[{"call": "A"}, {"call": "B"}], {"sleep": {"SleepCommand": [{"Load": {"Min": 0, "Max": 100}, "type": "static", "data": {"time": "10ms"}}]}}]`),
			Script{
				ConcurrentCommand{
					RequestCommand{ServiceName: "A"},
					RequestCommand{ServiceName: "B"},
				},
				SleepCommand{[]SleepCommandData{SleepCommandData{Range{uint64(0),uint64(100)}, Static, SleepCommandStatic{10 * time.Millisecond}}}},
			},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var script Script
			err := json.Unmarshal(test.input, &script)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}
			if !reflect.DeepEqual(test.script, script) {
				t.Errorf("expected %v; actual %v", test.script, script)
			}
		})
	}
}

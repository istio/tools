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

	"gonum.org/v1/gonum/stat/distuv"
)

func TestConcurrentCommand_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input   []byte
		command ConcurrentCommand
		err     error
	}{
		{
			[]byte(`[]`),
			ConcurrentCommand{},
			nil,
		},
		{
			[]byte(`[{"sleep": {"SleepCommand": [{"Load": {"Min": 0, "Max": 100}, "type": "static", "data": {"time": "1s"}}]}}]`),
			ConcurrentCommand{
				SleepCommand{[]SleepCommandData{SleepCommandData{Range{uint64(0),uint64(100)}, Static, SleepCommandStatic{1 * time.Second}}}},
			},
			nil,
		},
		{
			[]byte(`[{"call": "A"}, {"sleep": {"SleepCommand": [{"Load": {"Min": 0, "Max": 100}, "type": "static", "data": {"time": "10ms"}}]}}]`),
			ConcurrentCommand{
				RequestCommand{ServiceName: "A"},
				SleepCommand{[]SleepCommandData{SleepCommandData{Range{uint64(0),uint64(100)}, Static, SleepCommandStatic{10 * time.Millisecond}}}},
			},
			nil,
		},
		{
			[]byte(`[{"call": "A"}, {"sleep": {"SleepCommand": [{"Load": {"Min": 0, "Max": 100}, "type":"dist","Data":{"name":"normal", "Dist": {"Mu":1.0, "Sigma":0.25}}}]}}]`),
			ConcurrentCommand{
				RequestCommand{ServiceName: "A"},
				SleepCommand{[]SleepCommandData{SleepCommandData{Range{uint64(0),uint64(100)}, Distribution, SleepCommandDistribution{"normal", distuv.Normal{Mu: 1.0, Sigma: 0.25}}}}},
			},
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var command ConcurrentCommand
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

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
	"github.com/jmcvetta/randutil"
	"reflect"
	"testing"
	"time"
)

func TestSleepCommand_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input   []byte
		command SleepCommand
		err     error
	}{
		{
			[]byte(`{"100ms": 100}`),
			SleepCommand([]randutil.Choice{{100, 100 * time.Millisecond}}),
			nil,
		},
	}

	for _, test := range tests {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			var probDistribution map[string]int
			err := json.Unmarshal(test.input, &probDistribution)
			if test.err != err {
				t.Errorf("expected %v; actual %v", test.err, err)
			}

			command := make(SleepCommand, 0, len(probDistribution))

			for timeString, percentage := range probDistribution {
				duration, _ := time.ParseDuration(timeString)

				command = append(command, randutil.Choice{percentage, duration})
			}

			eq := reflect.DeepEqual(test.command, command)
			if !eq {
				t.Errorf("expected %v; actual %v", test.command, command)
			}
		})
	}
}

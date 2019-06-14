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
	"time"
	"github.com/jmcvetta/randutil"
	"istio.io/fortio/log"
	"strconv"
)

// SleepCommand describes a command to pause for a duration.
type SleepCommand []randutil.Choice

// UnmarshalJSON converts a JSON object to a SleepCommand.
func (c *SleepCommand) UnmarshalJSON(b []byte) (err error) {
	var probDistribution map[string]int

	err = json.Unmarshal(b, &probDistribution)
	if err != nil {
		return
	}

	ret := make(SleepCommand, 0, len(probDistribution))
	totalPercentage := 0

	for timeString, percentage := range probDistribution {
		duration, err := time.ParseDuration(timeString)

		if err != nil {
			return err
		}

		
		ret = append(ret, randutil.Choice{percentage, duration})
		totalPercentage += percentage
	}

	if totalPercentage != 100 {
		log.Fatalf("Total Percentage does not equal 100.")
	}

	*c = ret
	return
}

func (c SleepCommand) String() string {
	ret := ": "
	for idx, item := range c {
		if idx == 0 {
			ret += "{"
		}

		ret += "'" + item.Item.(time.Duration).String() + "': "
		ret += (strconv.Itoa(item.Weight))

		if (idx + 1) < len(c) {
			ret += ", "
		} else if (idx + 1) == len(c) {
			ret += "}"
		}

		
	}

	return ret
}

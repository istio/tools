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

	"gonum.org/v1/gonum/stat/distuv"
)

// ProcessCommand describes a command which contains a probability distribution
// of durations that the process should pause for.
type ProcessCommandConfig struct {
	Dist  string  `json:"dist"`
	Mean  float64 `json:"mean"`
	Sigma float64 `json:"sigma"`
}

type ProcessCommand struct {
	Dist interface {
		Rand() float64
	}
}

// UnmarshalJSON converts a JSON object to a ProcessCommand.
func (c *ProcessCommand) UnmarshalJSON(b []byte) (err error) {
	var data ProcessCommandConfig

	err = json.Unmarshal(b, &data)
	if err != nil {
		return
	}

	if data.Dist == "Normal" {
		dist := distuv.Normal{
			Mu:    data.Mean,
			Sigma: data.Sigma,
		}

		*c = ProcessCommand{Dist: dist}
	} else if data.Dist == "LogNormal" {
		dist := distuv.LogNormal{
			Mu:    data.Mean,
			Sigma: data.Sigma,
		}

		*c = ProcessCommand{Dist: dist}
	}

	return
}

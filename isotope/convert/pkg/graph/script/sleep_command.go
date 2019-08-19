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
	"log"
	"math/rand"
	"time"

	"github.com/jmcvetta/randutil"
	"gonum.org/v1/gonum/stat/distuv"
)

type CommandType string

const (
	Static       CommandType = "static"
	Histogram    CommandType = "histogram"
	Distribution CommandType = "dist"
	RawData      CommandType = "raw"
)

type SleepCommandRaw struct {
	List []float64 `json:"list"`
}

func (c SleepCommandRaw) Duration() time.Duration {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	return (time.Duration(c.List[rand.Intn(len(c.List))]))
}

type SleepCommandStatic struct {
	Time time.Duration `json:"time"`
}

func (c SleepCommandStatic) Duration() time.Duration {
	return c.Time * 1e6
}

type SleepCommandDistribution struct {
	DistName string `json:"name"`
	Dist     interface {
		Rand() float64
	} `json:"Dist"`
}

func (c SleepCommandDistribution) Duration() time.Duration {
	return time.Duration(c.Dist.Rand() * 1e6)
}

type SleepCommandHistogram struct {
	Histogram []randutil.Choice
}

func (c SleepCommandHistogram) Duration() time.Duration {
	result, err := randutil.WeightedChoice(c.Histogram)
	if err != nil {
		panic(err)
	}

	return (result.Item.(time.Duration) * 1e6)
}

type Range struct {
	Minimum uint64 `json:"Min"`
	Maximum uint64 `json:"Max"`
}

type SleepCommandWrapper struct {
	Load Range           `json:"Load"`
	Type CommandType     `json:"Type"`
	Data json.RawMessage `json:"Data"`
}

type SleepCommand struct {
	SleepCommand []SleepCommandData
}

type SleepCommandData struct {
	Load Range `json:"Load"`
	Type CommandType
	Data interface {
		Duration() time.Duration
	}
}

// UnmarshalJSON converts a JSON object to a SleepCommand.
func (c *SleepCommand) UnmarshalJSON(b []byte) (err error) {
	isJSONString := b[0] == '"'

	if isJSONString {
		var s string
		err = json.Unmarshal(b, &s)
		if err != nil {
			return
		}

		b = []byte(s)
	}

	var commands map[string][]SleepCommandWrapper
	err = json.Unmarshal(b, &commands)
	if err != nil {
		return
	}

	for _, command := range commands["SleepCommand"] {
		switch command.Type {
		case Static:
			var cmd map[string]interface{}
			err = json.Unmarshal(command.Data, &cmd)

			if err != nil {
				return
			}

			var staticCmd SleepCommandStatic

			t := cmd["time"]
			switch t.(type) {
			case string:
				staticCmd.Time, err = time.ParseDuration(cmd["time"].(string))

				if err != nil {
					return
				}
			case float64:
				staticCmd.Time = time.Duration(cmd["time"].(float64)) * time.Nanosecond
			}

			c.SleepCommand = append(c.SleepCommand, SleepCommandData{command.Load, command.Type, staticCmd})

		case Distribution:
			var cmd map[string]interface{}
			err = json.Unmarshal(command.Data, &cmd)

			if err != nil {
				return
			}

			var distCmd SleepCommandDistribution
			var dist interface {
				Rand() float64
			}

			switch cmd["name"] {
			case "normal":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.Normal{
					Mu:    distData["Mu"].(float64),
					Sigma: distData["Sigma"].(float64),
				}
			case "lognormal":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.LogNormal{
					Mu:    distData["Mu"].(float64),
					Sigma: distData["Sigma"].(float64),
				}
			case "beta":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.Beta{
					Alpha: distData["Alpha"].(float64),
					Beta:  distData["Beta"].(float64),
				}
			case "chi-squared":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.ChiSquared{
					K: distData["K"].(float64),
				}
			case "exp":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.Exponential{
					Rate: distData["Rate"].(float64),
				}
			case "f":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.F{
					D1: distData["D1"].(float64),
					D2: distData["D2"].(float64),
				}
			case "gamma":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.Gamma{
					Alpha: distData["Alpha"].(float64),
					Beta:  distData["Beta"].(float64),
				}
			case "gumbel-right":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.GumbelRight{
					Mu:   distData["Mu"].(float64),
					Beta: distData["Beta"].(float64),
				}
			case "inverse-gamma":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.InverseGamma{
					Alpha: distData["Alpha"].(float64),
					Beta:  distData["Beta"].(float64),
				}
			case "laplace":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.Laplace{
					Mu:    distData["Mu"].(float64),
					Scale: distData["Scale"].(float64),
				}
			case "pareto":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.Pareto{
					Xm:    distData["Xm"].(float64),
					Alpha: distData["Alpha"].(float64),
				}
			case "studentst":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.StudentsT{
					Mu:    distData["Mu"].(float64),
					Sigma: distData["Sigma"].(float64),
					Nu:    distData["Nu"].(float64),
				}
			case "weibull":
				distData := cmd["Dist"].(map[string]interface{})
				dist = distuv.Weibull{
					K:      distData["K"].(float64),
					Lambda: distData["Lambda"].(float64),
				}
			}

			distCmd.DistName = cmd["name"].(string)
			distCmd.Dist = dist
			c.SleepCommand = append(c.SleepCommand, SleepCommandData{command.Load, command.Type, distCmd})

		case Histogram:
			var probDistribution map[string]int
			err = json.Unmarshal(command.Data, &probDistribution)

			if err != nil {
				return
			}

			ret := make([]randutil.Choice, 0, len(probDistribution))
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

			var HistCmd SleepCommandHistogram
			HistCmd.Histogram = ret
			c.SleepCommand = append(c.SleepCommand, SleepCommandData{command.Load, command.Type, HistCmd})
		case RawData:
			var rawData SleepCommandRaw
			err = json.Unmarshal(command.Data, &rawData)

			if err != nil {
				return
			}

			c.SleepCommand = append(c.SleepCommand, SleepCommandData{command.Load, command.Type, rawData})
		}
	}
	return
}

func (c SleepCommand) String() string {
	res, err := json.Marshal(c)

	if err != nil {
		log.Fatalf("%s", err)
	}

	return string(res)
}

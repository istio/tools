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
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmcvetta/randutil"
	"gonum.org/v1/gonum/stat/distuv"
)

type SleepCommandStatic struct {
	time time.Duration `json:"time"`
}

func (c SleepCommandStatic) Duration() time.Duration {
	return c.time
}

type SleepCommandDistribution struct {
	Type string
	Dist interface {
		Rand() float64
	}
}

func (c SleepCommandDistribution) Duration() time.Duration {
	return time.Duration(c.Dist.Rand() * 10e8)
}

type SleepCommandHistogram struct {
	Histogram []randutil.Choice
}

func (c SleepCommandHistogram) Duration() time.Duration {
	result, err := randutil.WeightedChoice(c.Histogram)
	if err != nil {
		panic(err)
	}

	return (time.Duration(result.Item.(time.Duration)))
}

type SleepCommandWrapper struct {
	Type string `json:"type"`
	Data json.RawMessage
}

type SleepCommand struct {
	Type string
	Data interface {
		Duration() time.Duration
	}
}

// UnmarshalJSON converts a JSON object to a SleepCommand.
func (c *SleepCommand) UnmarshalJSON(b []byte) (err error) {
	var command SleepCommandWrapper
	err = json.Unmarshal(b, &command)
	if err != nil {
		return
	}

	switch command.Type {
	case "static":
		var cmd map[string]interface{}
		err = json.Unmarshal(command.Data, &cmd)

		if err != nil {
			return
		}

		var staticCmd SleepCommandStatic
		staticCmd.time, err = time.ParseDuration(cmd["time"].(string))

		if err != nil {
			return
		}

		*c = SleepCommand{command.Type, staticCmd}

	case "dist":
		var cmd map[string]interface{}
		err = json.Unmarshal(command.Data, &cmd)

		if err != nil {
			return
		}

		switch cmd["dist"] {
		case "normal":
			dist := distuv.Normal{
				Mu:    cmd["mean"].(float64),
				Sigma: cmd["sigma"].(float64),
			}

			var distCmd SleepCommandDistribution
			distCmd.Type = cmd["dist"].(string)
			distCmd.Dist = dist
			*c = SleepCommand{command.Type, distCmd}
		case "lognormal":
			dist := distuv.LogNormal{
				Mu:    cmd["mean"].(float64),
				Sigma: cmd["sigma"].(float64),
			}

			var distCmd SleepCommandDistribution
			distCmd.Type = cmd["dist"].(string)
			distCmd.Dist = dist
			*c = SleepCommand{command.Type, distCmd}
		}
	case "histogram":
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
		*c = SleepCommand{command.Type, HistCmd}

	}
	return
}

func (c SleepCommand) String() string {
	switch c.Type {
	case "static":
		return time.Duration(c.Data.Duration()).String()
	case "dist":
		dist := c.Data.(SleepCommandDistribution)

		switch dist.Type {
		case "normal":
			mean := fmt.Sprintf("%f", (dist.Dist.(distuv.Normal).Mu))
			sigma := fmt.Sprintf("%f", (dist.Dist.(distuv.Normal).Sigma))
			return fmt.Sprintf("Distribution: Normal {Mean: %s, Sigma: %s}", mean, sigma)
		case "lognormal":
			mean := fmt.Sprintf("%f", (dist.Dist.(distuv.LogNormal).Mu))
			sigma := fmt.Sprintf("%f", (dist.Dist.(distuv.LogNormal).Sigma))
			return fmt.Sprintf("Distribution: LogNormal {Mean: %s, Sigma: %s}", mean, sigma)
		}

	case "histogram":
		Histogram := c.Data.(SleepCommandHistogram).Histogram
		var str strings.Builder

		str.WriteString("Histogram: ")

		for idx, item := range Histogram {
			// prob := item.Weight
			duration := item.Item.(time.Duration)
			weight := item.Weight

			str.WriteString("{")
			str.WriteString(strconv.Itoa(weight))
			str.WriteString(",")
			str.WriteString(duration.String())
			str.WriteString("}")

			if idx+1 < len(Histogram) {
				str.WriteString(" ")
			}
		}

		return str.String()
	}

	return "Error: Incorrect sleep command type"
}

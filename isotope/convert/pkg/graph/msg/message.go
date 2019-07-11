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

package msg

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/jmcvetta/randutil"
	"gonum.org/v1/gonum/stat/distuv"

	"istio.io/tools/isotope/convert/pkg/graph/size"
)

type CommandType string

const (
	Static       CommandType = "static"
	Histogram    CommandType = "histogram"
	Distribution CommandType = "dist"
)

// MessageSizeStatic Type and Associated Functions
type MessageSizeStatic struct {
	Length size.ByteSize `json:"size"`
	Number int           `json:"number"`
}

func (c MessageSizeStatic) Size() size.ByteSize {
	return c.Length
}

func (c MessageSizeStatic) Amount() int {
	return c.Number
}

// MessageSizeDistribution and Associated Functions
type MessageSizeDistribution struct {
	Type   string
	Unit   string
	Number int
	Dist   interface {
		Rand() float64
	}
}

func FloatToString(inputNum float64) string {
	return strconv.FormatFloat(inputNum, 'f', 10, 64)
}

func (c MessageSizeDistribution) Size() size.ByteSize {
	var str strings.Builder

	str.WriteString(FloatToString(c.Dist.Rand()))
	str.WriteString(c.Unit)

	result, err := size.FromString(str.String())
	if err != nil {
		panic(err)
	}

	return result
}

func (c MessageSizeDistribution) Amount() int {
	return c.Number
}

// MessageSizeHistogram Type and Associated Functions
type MessageSizeHistogram struct {
	Histogram []randutil.Choice
	Number    int `json:"number"`
}

func (c MessageSizeHistogram) Size() size.ByteSize {
	result, err := randutil.WeightedChoice(c.Histogram)
	if err != nil {
		panic(err)
	}

	return (result.Item.(size.ByteSize))
}

func (c MessageSizeHistogram) Amount() int {
	return c.Number
}

type MessageSize struct {
	Type CommandType `json:"type"`
	Data interface {
		Size() size.ByteSize
		Amount() int
	} `json:"data"`
}

type MessageSizeWrapper struct {
	Type CommandType `json:"type"`
	Data json.RawMessage
}

func (c *MessageSize) UnmarshalJSON(b []byte) (err error) {
	var command MessageSizeWrapper
	err = json.Unmarshal(b, &command)
	if err != nil {
		return
	}

	switch command.Type {
	case Static:
		var cmd map[string]interface{}
		err = json.Unmarshal(command.Data, &cmd)

		if err != nil {
			return
		}

		var staticCmd MessageSizeStatic
		staticCmd.Length, err = size.FromString(cmd["size"].(string))
		if err != nil {
			return
		}
		staticCmd.Number = int(cmd["number"].(float64))

		*c = MessageSize{command.Type, staticCmd}

	case Distribution:
		var cmd map[string]interface{}
		err = json.Unmarshal(command.Data, &cmd)

		if err != nil {
			return
		}

		var distCmd MessageSizeDistribution
		distCmd.Number = int(cmd["number"].(float64))
		distCmd.Unit = cmd["unit"].(string)
		var dist interface {
			Rand() float64
		}

		switch cmd["distribution"] {
		case "normal":
			dist = distuv.Normal{
				Mu:    cmd["mean"].(float64),
				Sigma: cmd["sigma"].(float64),
			}
		case "lognormal":
			dist = distuv.LogNormal{
				Mu:    cmd["mean"].(float64),
				Sigma: cmd["sigma"].(float64),
			}
		case "beta":
			dist = distuv.Beta{
				Alpha: cmd["alpha"].(float64),
				Beta:  cmd["beta"].(float64),
			}
		case "chi-squared":
			dist = distuv.ChiSquared{
				K: cmd["k"].(float64),
			}
		case "exp":
			dist = distuv.Exponential{
				Rate: cmd["rate"].(float64),
			}
		case "f":
			dist = distuv.F{
				D1: cmd["d1"].(float64),
				D2: cmd["d2"].(float64),
			}
		case "gamma":
			dist = distuv.Gamma{
				Alpha: cmd["alpha"].(float64),
				Beta:  cmd["beta"].(float64),
			}
		case "gumbel-right":
			dist = distuv.GumbelRight{
				Mu:   cmd["mu"].(float64),
				Beta: cmd["beta"].(float64),
			}
		case "inverse-gamma":
			dist = distuv.InverseGamma{
				Alpha: cmd["alpha"].(float64),
				Beta:  cmd["beta"].(float64),
			}
		case "laplace":
			dist = distuv.Laplace{
				Mu:    cmd["mu"].(float64),
				Scale: cmd["scale"].(float64),
			}
		case "pareto":
			dist = distuv.Pareto{
				Xm:    cmd["xm"].(float64),
				Alpha: cmd["alpha"].(float64),
			}
		case "studentst":
			dist = distuv.StudentsT{
				Mu:    cmd["mu"].(float64),
				Sigma: cmd["sigma"].(float64),
				Nu:    cmd["nu"].(float64),
			}
		case "weibull":
			dist = distuv.Weibull{
				K:      cmd["k"].(float64),
				Lambda: cmd["lambda"].(float64),
			}
		}
		distCmd.Type = cmd["distribution"].(string)
		distCmd.Dist = dist
		*c = MessageSize{command.Type, distCmd}

	case Histogram:
		var cmd map[string]interface{}
		err = json.Unmarshal(command.Data, &cmd)

		if err != nil {
			return
		}

		var probDistribution = cmd["histogram"].(map[string]interface{})
		ret := make([]randutil.Choice, 0, len(probDistribution))
		totalPercentage := 0

		for sizeString, percentage := range probDistribution {
			percentage := int(percentage.(float64))
			size, err := size.FromString(sizeString)

			if err != nil {
				return err
			}

			ret = append(ret, randutil.Choice{percentage, size})
			totalPercentage += percentage
		}

		if totalPercentage != 100 {
			log.Fatalf("Total Percentage does not equal 100.")
		}

		var HistCmd MessageSizeHistogram
		HistCmd.Number = int(cmd["number"].(float64))
		HistCmd.Histogram = ret
		*c = MessageSize{command.Type, HistCmd}
	}
	return
}

func (c MessageSize) String() string {
	res, err := json.Marshal(c)

	if err != nil {
		log.Fatalf("%s", err)
	}

	return string(res)
}

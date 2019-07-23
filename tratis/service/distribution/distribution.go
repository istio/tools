// Copyright 2019 Istio Authors
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

package distribution

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type Details struct {
	Name       string    `json:"name"`
	Parameters []float64 `json:"parameters"`
	Mean       float64   `json:"mean"`
	Sigma      float64   `json:"sigma"`
}

type TotalDistributions struct {
	OperationName string    `json:"operationName"`
	Distributions []Details `json:"distributions"`
}

type HasDistributionData interface {
	GetDistributionData() [][]uint64
	GetOperation() string
	Convert() HasDistributionData
}

func ConvertTimeInfo(data []CombinedTimeInformation) []HasDistributionData {
	ret := make([]HasDistributionData, 0)

	for _, item := range data {
		ret = append(ret, HasDistributionData(item))
	}

	return ret
}

func ConvertSizeInfo(data []CombinedSizeInformation) []HasDistributionData {
	ret := make([]HasDistributionData, 0)

	for _, item := range data {
		ret = append(ret, HasDistributionData(item))
	}

	return ret
}

func InfoToDist(fileName string,
	funcName string, data []HasDistributionData) []TotalDistributions {
	ret := make([]TotalDistributions, len(data))

	for idx, distributionData := range data {
		for _, data := range distributionData.GetDistributionData() {
			cmd := GeneratePythonCommand(fileName, funcName, data)
			ret[idx].Distributions = append(ret[idx].Distributions, RunDistributionFitting(cmd))
		}

		ret[idx].OperationName = distributionData.GetOperation()
	}

	return ret
}

func GeneratePythonCommand(fileName string, funcName string, data []uint64) string {
	var command strings.Builder

	s := fmt.Sprintf("import %s; print(%s.%s([", fileName, fileName, funcName)
	command.WriteString(s)

	for idx, value := range data {
		command.WriteString(strconv.Itoa(int(value)))

		if idx+1 < len(data) {
			command.WriteString(", ")
		}
	}

	command.WriteString("]))")

	return command.String()
}

/*
The more values (traces) generated the better the results of the
distribution fitting module but at least a minimum of 50 values
(traces) should be generated.
*/

func RunDistributionFitting(command string) Details {
	cmd := exec.Command("python", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		fmt.Println(string(out))
	}

	var ret Details
	err = json.Unmarshal(out, &ret)
	if err != nil {
		log.Fatalf(`Python Script Output Not Correctly Formatted: %s`, err)
	}

	return ret
}

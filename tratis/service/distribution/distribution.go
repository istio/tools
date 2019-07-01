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

package distribution

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type DistributionDetails struct {
	Name       string    `json:"name"`
	Parameters []float64 `json:"parameters"`
}

type TotalDistributions struct {
	OperationName string                `json:"operationName"`
	Distributions []DistributionDetails `json:"distributions"`
}

func TimeInfoToDist(fileName string, funcName string, data []CombinedTimeInformation) {
	ret := make([]TotalDistributions, len(data))

	for idx, operation := range data {
		for _, data := range operation.Duration {
			cmd := GeneratePythonCommand(fileName, funcName, data)
			ret[idx].Distributions = append(ret[idx].Distributions, RunDistributionFitting(cmd))
		}

		ret[idx].OperationName = operation.OperationName
	}

	fmt.Println(ret)
}

func GeneratePythonCommand(fileName string, funcName string, data []uint64) string {
	var command strings.Builder
	command.WriteString("import ")
	command.WriteString(fileName)
	command.WriteString("; print ")
	command.WriteString(fileName)
	command.WriteString(".")
	command.WriteString(funcName)
	command.WriteString("([")

	for idx, value := range data {
		command.WriteString(strconv.Itoa(int(value)))

		if idx+1 < len(data) {
			command.WriteString(", ")
		}
	}

	command.WriteString("])")

	return command.String()
}

func RunDistributionFitting(command string) DistributionDetails {
	cmd := exec.Command("python", "-c", command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	var ret DistributionDetails
	err = json.Unmarshal(out, &ret)
	if err != nil {
		log.Fatalf(`Python Script Output Not Correctly Formatted`)
	}

	return ret
}

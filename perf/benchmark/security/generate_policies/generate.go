// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"

	authzpb "istio.io/api/security/v1beta1"
)

type generator interface {
	generate(action string, ruleMap map[string]int) *authzpb.Rule
}

type operationGenerator struct {
}

func (operationGenerator) generate(_ string, ruleMap map[string]int) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listOperation []*authzpb.Rule_To

	numPaths := ruleMap["numPaths"]
	if numPaths > 0 {
		paths := make([]string, numPaths)
		for i := 0; i < numPaths; i++ {
			paths[i] = fmt.Sprintf("/Invalid-path-%d", i)
		}
		operation := &authzpb.Rule_To{
			Operation: &authzpb.Operation{
				Paths: paths,
			},
		}
		listOperation = append(listOperation, operation)
	}
	rule.To = listOperation
	return rule
}

type conditionGenerator struct {
}

func (conditionGenerator) generate(action string, ruleMap map[string]int) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listCondition []*authzpb.Condition

	numValues := ruleMap["numValues"]
	if numValues > 0 {
		values := make([]string, numValues)
		for i := 0; i < numValues; i++ {
			if i == numValues - 1 && action == "ALLOW" {
				values[i] = "admin"
			} else {
				values[i] = "guest"
			}
		}
		condition := &authzpb.Condition{
			Key:    "request.headers[x-token]",
			Values: values,
		}
		listCondition = append(listCondition, condition)
	}
	rule.When = listCondition
	return rule
}

type sourceGenerator struct {
}

func (sourceGenerator) generate(_ string, ruleMap map[string]int) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listSource []*authzpb.Rule_From

	numSourceIP := ruleMap["numSourceIP"]
	if numSourceIP > 0 {
		sourceIPList := make([]string, numSourceIP)
		for i := 0; i < numSourceIP; i++ {
			sourceIPList[i] = fmt.Sprintf("0.0.%d.%d", i / 256, i % 256)
		}
		source := &authzpb.Rule_From{
			Source: &authzpb.Source{
				IpBlocks: sourceIPList,
			},
		}
		listSource = append(listSource, source)
	}

	numNamepaces := ruleMap["numNamespaces"]
	if numNamepaces > 0 {
		namespaces := make([]string, numNamepaces)
		for i := 0; i < numNamepaces; i++ {
			namespaces[i] = fmt.Sprintf("Invalid-namespace-%d", i)
		}
		source := &authzpb.Rule_From{
			Source: &authzpb.Source{
				Namespaces: namespaces,
			},
		}
		listSource = append(listSource, source)
	}
	rule.From = listSource
	return rule
}

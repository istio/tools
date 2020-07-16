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
	generate(kind string, num int, action string) *authzpb.Rule
}

type operationGenerator struct {
}

func (operationGenerator) generate(kind string, num int, _ string) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listOperation []*authzpb.Rule_To

	for i := 0; i < num; i++ {
		path := fmt.Sprintf("%s%d", "/invalid-path-", i)
		operation := &authzpb.Rule_To{
			Operation: &authzpb.Operation{
				Methods: []string{"GET", "HEAD"},
				Paths:   []string{path},
			},
		}
		listOperation = append(listOperation, operation)
	}
	rule.To = listOperation
	return rule
}

type conditionGenerator struct {
}

func (conditionGenerator) generate(kind string, num int, action string) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listCondition []*authzpb.Condition

	for i := 0; i < num; i++ {
		values := []string{"guest"}
		// Allow the last rule to match a request from "admin"
		if i == num-1 && action == "ALLOW" {
			values = []string{"admin"}
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

func (sourceGenerator) generate(kind string, num int, _ string) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listSource []*authzpb.Rule_From

	for i := 0; i < num; i++ {
		namespace := fmt.Sprintf("%s%d", "invalid-namespace-", i)
		source := &authzpb.Rule_From{
			Source: &authzpb.Source{
				Namespaces: []string{namespace},
			},
		}
		listSource = append(listSource, source)
	}
	rule.From = listSource
	return rule
}

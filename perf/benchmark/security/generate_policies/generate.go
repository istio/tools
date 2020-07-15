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
	authzpb "istio.io/api/security/v1beta1"
)

type generator interface {
	generate(kind string, num int, action string) (*authzpb.Rule, error)
}

type operationGenerator struct {
}

func (operationGenerator) generate(kind string, num int, _ string) (*authzpb.Rule, error) {
	rule := &authzpb.Rule{}
	var listOperation []*authzpb.Rule_To

	for i := 0; i < num; i++ {
		operation := &authzpb.Rule_To{}
		operation.Operation = &authzpb.Operation{}
		operation.Operation.Methods = []string{"GET", "HEAD"}
		operation.Operation.Paths = []string{"/invalid-path"}
		listOperation = append(listOperation, operation)
	}
	rule.To = listOperation
	return rule, nil
}

type conditionGenerator struct {
}

func (conditionGenerator) generate(kind string, num int, action string) (*authzpb.Rule, error) {
	rule := &authzpb.Rule{}
	var listCondition []*authzpb.Condition

	for i := 0; i < num; i++ {
		condition := &authzpb.Condition{}
		condition.Key = "request.headers[x-token]"
		// Sets the last rule to match
		if i == num-1 && action == "ALLOW" {
			condition.Values = []string{"admin"}
		} else {
			condition.Values = []string{"guest"}
		}
		listCondition = append(listCondition, condition)
	}
	rule.When = listCondition
	return rule, nil
}

type sourceGenerator struct {
}

func (sourceGenerator) generate(kind string, num int, _ string) (*authzpb.Rule, error) {
	rule := &authzpb.Rule{}
	var listSource []*authzpb.Rule_From

	for i := 0; i < num; i++ {
		source := &authzpb.Rule_From{}
		source.Source = &authzpb.Source{}
		source.Source.Namespaces = []string{"invalid-namespace"}
		listSource = append(listSource, source)
	}
	rule.From = listSource
	return rule, nil
}

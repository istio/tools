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
	generate(policyData SecurityPolicy) *authzpb.Rule
}

type operationGenerator struct{}

func (operationGenerator) generate(policyData SecurityPolicy) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listOperation []*authzpb.Rule_To

	if numPaths := policyData.AuthZ.NumPaths; numPaths > 0 {
		paths := make([]string, numPaths)
		for i := 0; i < numPaths; i++ {
			paths[i] = fmt.Sprintf("/invalid-path-%d", i)
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

type conditionGenerator struct{}

func (conditionGenerator) generate(policyData SecurityPolicy) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listCondition []*authzpb.Condition

	if numValues := policyData.AuthZ.NumValues; numValues > 0 {
		values := make([]string, numValues)
		for i := 0; i < numValues; i++ {
			if i == numValues-1 && policyData.AuthZ.Action == "ALLOW" {
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

type sourceGenerator struct{}

func (sourceGenerator) generate(policyData SecurityPolicy) *authzpb.Rule {
	rule := &authzpb.Rule{}
	var listSource []*authzpb.Rule_From

	if numSourceIP := policyData.AuthZ.NumSourceIP; numSourceIP > 0 {
		sourceIPList := make([]string, numSourceIP)
		for i := 0; i < numSourceIP; i++ {
			sourceIPList[i] = fmt.Sprintf("0.0.%d.%d", i/256, i%256)
		}
		source := &authzpb.Rule_From{
			Source: &authzpb.Source{
				IpBlocks: sourceIPList,
			},
		}
		listSource = append(listSource, source)
	}

	if numNamepaces := policyData.AuthZ.NumNamespaces; numNamepaces > 0 {
		namespaces := make([]string, numNamepaces)
		for i := 0; i < numNamepaces; i++ {
			namespaces[i] = fmt.Sprintf("invalid-namespace-%d", i)
		}
		source := &authzpb.Rule_From{
			Source: &authzpb.Source{
				Namespaces: namespaces,
			},
		}
		listSource = append(listSource, source)
	}

	if numPrincipals := policyData.AuthZ.NumPrincipals; numPrincipals > 0 {
		principals := make([]string, numPrincipals)
		for i := 0; i < numPrincipals; i++ {
			principals[i] = fmt.Sprintf("cluster.local/ns/twopods-istio/sa/Invalid-%d", i)
		}
		source := &authzpb.Rule_From{
			Source: &authzpb.Source{
				Principals: principals,
			},
		}
		listSource = append(listSource, source)
	}

	if numRequestPrincipals := policyData.AuthZ.NumRequestPrincipals; numRequestPrincipals > 0 {
		requestPrincipals := make([]string, numRequestPrincipals)
		for i := 0; i < numRequestPrincipals; i++ {
			principalValue := "invalid-issuer/subject"
			if i == numRequestPrincipals-1 {
				principalValue = fmt.Sprintf("issuer-%d/subject", policyData.RequestAuthN.NumJwks)
			}
			requestPrincipals[i] = principalValue
		}
		source := &authzpb.Rule_From{
			Source: &authzpb.Source{
				RequestPrincipals: requestPrincipals,
			},
		}
		listSource = append(listSource, source)
	}
	rule.From = listSource
	return rule
}

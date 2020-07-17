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
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"

	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/proto"

	"istio.io/istio/pkg/util/protomarshal"

	authzpb "istio.io/api/security/v1beta1"
)

type ruleOption struct {
	occurrence int
	gen        generator
}

type MyPolicy struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   MetadataStruct `json:"metadata"`
}

type MetadataStruct struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func ToYAML(policy *MyPolicy, spec proto.Message) (string, error) {
	header, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	headerYaml, err := yaml.JSONToYAML(header)
	if err != nil {
		return "", err
	}

	authorizationPolicy, err := protomarshal.ToYAML(spec)
	if err != nil {
		return "", err
	}

	rulesYaml := bytes.Buffer{}
	rulesYaml.WriteString("spec:\n")
	scanner := bufio.NewScanner(strings.NewReader(authorizationPolicy))
	for scanner.Scan() {
		rulesYaml.WriteString(" " + scanner.Text() + "\n")
	}
	return string(headerYaml) + rulesYaml.String(), nil
}

func getOrderedKeySlice(ruleToOccurrences map[string]*ruleOption) *[]string {
	var sortedKeys []string
	for key := range ruleToOccurrences {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	return &sortedKeys
}

func generateAuthorizationPolicy(action string, ruleToOccurrences map[string]*ruleOption, policy *MyPolicy) (string, error) {
	spec := &authzpb.AuthorizationPolicy{}
	switch action {
	case "ALLOW":
		spec.Action = authzpb.AuthorizationPolicy_ALLOW
	case "DENY":
		spec.Action = authzpb.AuthorizationPolicy_DENY
	}

	var ruleList []*authzpb.Rule
	sortedKeys := getOrderedKeySlice(ruleToOccurrences)
	for _, name := range *sortedKeys {
		ruleOp := ruleToOccurrences[name]
		rule := ruleOp.gen.generate(name, ruleOp.occurrence, action)
		ruleList = append(ruleList, rule)
	}
	spec.Rules = ruleList

	yaml, err := ToYAML(policy, spec)
	if err != nil {
		return "", err
	}
	return yaml, nil
}

func generateRule(action string, ruleToOccurrences map[string]*ruleOption,
	policy *MyPolicy) (string, error) {

	switch policy.Kind {
	case "AuthorizationPolicy":
		return generateAuthorizationPolicy(action, ruleToOccurrences, policy)
	case "PeerAuthentication":
		return "", fmt.Errorf("unimplemented")
	case "RequestAuthentication":
		return "", fmt.Errorf("unimplemented")
	default:
		return "", fmt.Errorf("invalid policy")
	}
}

func createRules(action string, ruleToOccurrences map[string]*ruleOption, policy *MyPolicy) (string, error) {
	yaml, err := generateRule(action, ruleToOccurrences, policy)
	if err != nil {
		return "", err
	}
	return yaml, nil
}

func createPolicyHeader(namespace string, name string, kind string) *MyPolicy {
	metadata := MetadataStruct{namespace, name}
	return &MyPolicy{
		APIVersion: "security.istio.io/v1beta1",
		Kind:       kind,
		Metadata:   metadata,
	}
}

func createRuleOptionMap(ruleToOccurancesPtr map[string]*int) *map[string]*ruleOption {
	ruleOptionMap := make(map[string]*ruleOption)
	for rule, occurrence := range ruleToOccurancesPtr {
		ruleOptionMap[rule] = &ruleOption{}
		ruleOptionMap[rule].occurrence = *occurrence
		switch rule {
		case "when":
			ruleOptionMap[rule].gen = conditionGenerator{}
		case "to":
			ruleOptionMap[rule].gen = operationGenerator{}
		case "from":
			ruleOptionMap[rule].gen = sourceGenerator{}
		default:
			fmt.Println("invalid rules")
		}
	}
	return &ruleOptionMap
}

func main() {
	namespacePtr := flag.String("namespace", "twopods-istio", "Namespace in which the rule shall be applied to")
	policyType := flag.String("policyType", "AuthorizationPolicy", "The type of security policy")
	actionPtr := flag.String("action", "DENY", "Type of action")
	numPoliciesPtr := flag.Int("numPolicies", 1, "Number of policies wanted")

	ruleToOccurancesPtr := make(map[string]*int)
	ruleToOccurancesPtr["when"] = flag.Int("when", 1, "Number of when condition wanted")
	ruleToOccurancesPtr["to"] = flag.Int("to", 1, "Number of To operations wanted")
	ruleToOccurancesPtr["from"] = flag.Int("from", 1, "Number of From sources wanted")
	flag.Parse()

	for i := 1; i <= *numPoliciesPtr; i++ {
		yaml := bytes.Buffer{}
		policy := createPolicyHeader(fmt.Sprintf("%s%d", "test-", i), *namespacePtr, *policyType)

		ruleOptionMap := createRuleOptionMap(ruleToOccurancesPtr)
		if rules, err := createRules(*actionPtr, *ruleOptionMap, policy); err != nil {
			fmt.Println(err)
		} else {
			yaml.WriteString(rules)
			fmt.Println(yaml.String())
		}
	}
}

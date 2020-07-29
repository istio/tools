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
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

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

func ToJSON(msg proto.Message) (string, error) {
	return ToJSONWithIndent(msg, "")
}

func ToJSONWithIndent(msg proto.Message, indent string) (string, error) {
	if msg == nil {
		return "", fmt.Errorf("unexpected nil message")
	}

	m := jsonpb.Marshaler{Indent: indent}
	return m.MarshalToString(msg)
}

func ToYAML(msg proto.Message) (string, error) {
	js, err := ToJSON(msg)
	if err != nil {
		return "", err
	}
	yml, err := yaml.JSONToYAML([]byte(js))
	return string(yml), err
}

func PolicyToYAML(policy *MyPolicy, spec proto.Message) (string, error) {
	header, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}

	headerYaml, err := yaml.JSONToYAML(header)
	if err != nil {
		return "", err
	}

	authorizationPolicy, err := ToYAML(spec)
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
		if ruleOp.occurrence > 0 {
			rule := ruleOp.gen.generate(name, ruleOp.occurrence, action)
			ruleList = append(ruleList, rule)
		}
	}
	spec.Rules = ruleList

	yaml, err := PolicyToYAML(policy, spec)
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
		return "", fmt.Errorf("unknown policy kind: %s", policy.Kind)
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
	return &MyPolicy{
		APIVersion: "security.istio.io/v1beta1",
		Kind:       kind,
		Metadata:   MetadataStruct{Namespace: namespace, Name: name},
	}
}

func createRuleOptionMap(ruleToOccurancesPtr map[string]int) (map[string]*ruleOption, error) {
	ruleOptionMap := make(map[string]*ruleOption)
	for rule, occurrence := range ruleToOccurancesPtr {
		ruleOptionMap[rule] = &ruleOption{}
		ruleOptionMap[rule].occurrence = occurrence
		switch rule {
		case "when":
			ruleOptionMap[rule].gen = conditionGenerator{}
		case "to":
			ruleOptionMap[rule].gen = operationGenerator{}
		case "from":
			ruleOptionMap[rule].gen = sourceGenerator{}
		default:
			return nil, fmt.Errorf("invalid rule: %s", rule)
		}
	}
	return ruleOptionMap, nil
}

func parseArguments(arguments string) map[string]string {
	argumentMap := make(map[string]string)
	for _, arg := range strings.Split(arguments, ",") {
		keyValue := strings.Split(arg, ":")
		argumentMap[keyValue[0]] = keyValue[1];
	}
	return argumentMap
}

func parseHeader(arguments map[string]string) map[string]string {
	headerMap := make(map[string]string)
	// These are the default values
	headerMap["namespace"] = "twopods-istio"
	headerMap["policyType"] = "AuthorizationPolicy"
	headerMap["action"] = "DENY"
	headerMap["numPolicies"] = "1"

	for key := range headerMap {
		if argVal, inMap  := arguments[key]; inMap {
			headerMap[key] = argVal
		}
	}
	return headerMap;
}

func parseRules(arguments map[string]string) (map[string]int, error) {
	ruleMap := make(map[string]int)
	// These are the default values
	ruleMap["when"] = 0
	ruleMap["from"] = 1
	ruleMap["to"] = 0

	for key := range ruleMap {
		if argVal, inMap := arguments[key]; inMap {
			argVal, err := strconv.Atoi(argVal)
			if err != nil {
				return nil, fmt.Errorf("invalid value: %s", ruleMap["numPolicies"])
			}
			ruleMap[key] = argVal
		}
	}
	return ruleMap, nil
}

func main() {
	securityPtr := flag.String("security_option", "numPolicies:1", "List of key value pairs seperated by commas")
	flag.Parse()

	argumentMap := parseArguments(*securityPtr)
	headerMap := parseHeader(argumentMap)
	ruleMap, err := parseRules(argumentMap)
	if err != nil {
		fmt.Println(err)
		return
	}

	numPolices, err := strconv.Atoi(headerMap["numPolicies"])
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 1; i <= numPolices; i++ {
		yaml := bytes.Buffer{}
		policy := createPolicyHeader(headerMap["namespace"], fmt.Sprintf("test-%d", i), headerMap["policyType"])

		ruleOptionMap, err := createRuleOptionMap(ruleMap)
		if err != nil {
			fmt.Println(err)
			break
		}

		rules, err := createRules(headerMap["action"], ruleOptionMap, policy)
		if err != nil {
			fmt.Println(err)
			break
		} else {
			yaml.WriteString(rules)
			if i < numPolices {
				yaml.WriteString("---")
			}
			fmt.Println(yaml.String())
		}
	}
}

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

type ruleGenerator struct {
	gen generator
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

func getOrderedKeySlice(ruleToGenerator map[string]*ruleGenerator) *[]string {
	var sortedKeys []string
	for key := range ruleToGenerator {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	return &sortedKeys
}

func generateAuthorizationPolicy(action string, ruleToGenerator map[string]*ruleGenerator, policy *MyPolicy,
	ruleMap map[string]int) (string, error) {
	spec := &authzpb.AuthorizationPolicy{}
	switch action {
	case "ALLOW":
		spec.Action = authzpb.AuthorizationPolicy_ALLOW
	case "DENY":
		spec.Action = authzpb.AuthorizationPolicy_DENY
	}

	var ruleList []*authzpb.Rule
	sortedKeys := getOrderedKeySlice(ruleToGenerator)
	for _, name := range *sortedKeys {
		rule := ruleToGenerator[name].gen.generate(action, ruleMap)
		ruleList = append(ruleList, rule)
	}
	spec.Rules = ruleList

	yaml, err := PolicyToYAML(policy, spec)
	if err != nil {
		return "", err
	}
	return yaml, nil
}

func generatePeerAuthentication(mtlsMode string, policy *MyPolicy) (string, error) {
	spec := &authzpb.PeerAuthentication{
		Mtls: &authzpb.PeerAuthentication_MutualTLS{},
	}
	switch mtlsMode {
	case "STRICT":
		spec.Mtls.Mode = authzpb.PeerAuthentication_MutualTLS_STRICT
	case "DISABLE":
		spec.Mtls.Mode = authzpb.PeerAuthentication_MutualTLS_DISABLE
	default:
		return "", fmt.Errorf("invalid mtlsMode: %s", mtlsMode)
	}

	yaml, err := PolicyToYAML(policy, spec)
	if err != nil {
		return "", err
	}
	return yaml, nil
}

func generateRules(argumentMap map[string]string, ruleToGenerator map[string]*ruleGenerator,
	policy *MyPolicy, ruleMap map[string]int) (string, error) {

	switch policy.Kind {
	case "AuthorizationPolicy":
		return generateAuthorizationPolicy(argumentMap["action"], ruleToGenerator, policy, ruleMap)
	case "PeerAuthentication":
		return generatePeerAuthentication(argumentMap["mtlsMode"], policy)
	case "RequestAuthentication":
		return "", fmt.Errorf("unimplemented")
	default:
		return "", fmt.Errorf("unknown policy kind: %s", policy.Kind)
	}
}

func createPolicyHeader(namespace string, name string, kind string) *MyPolicy {
	return &MyPolicy{
		APIVersion: "security.istio.io/v1beta1",
		Kind:       kind,
		Metadata:   MetadataStruct{Namespace: namespace, Name: name},
	}
}

func createRuleGeneratorMap(ruleToOccurancesPtr map[string]int) map[string]*ruleGenerator {
	ruleGeneratorMap := make(map[string]*ruleGenerator)

	if ruleToOccurancesPtr["numSourceIP"] > 0 || ruleToOccurancesPtr["numNamespaces"] > 0 ||
		ruleToOccurancesPtr["numPrincipals"] > 0 {
		ruleGeneratorMap["from"] = &ruleGenerator{
			gen: sourceGenerator{},
		}
	}

	if ruleToOccurancesPtr["numPaths"] > 0 {
		ruleGeneratorMap["to"] = &ruleGenerator{
			gen: operationGenerator{},
		}
	}

	if ruleToOccurancesPtr["numValues"] > 0 {
		ruleGeneratorMap["when"] = &ruleGenerator{
			gen: conditionGenerator{},
		}
	}
	return ruleGeneratorMap
}

func parseArguments(arguments string) (map[string]string, error) {
	argumentMap := make(map[string]string)
	// These are the default values
	argumentMap["namespace"] = "twopods-istio"
	argumentMap["action"] = "DENY"
	argumentMap["mtlsMode"] = "STRICT"

	if len(arguments) > 0 {
		for _, arg := range strings.Split(arguments, ",") {
			keyValue := strings.Split(arg, ":")
			if len(keyValue) == 1 {
				return nil, fmt.Errorf("invalid argument: %s", keyValue[0])
			}
			argumentMap[keyValue[0]] = keyValue[1]
		}
	}
	return argumentMap, nil
}

func createPolicyMap(argument map[string]string) (map[string]int, error) {
	policyMap := make(map[string]int)
	// These are the default values
	policyMap["AuthorizationPolicy"] = 0
	policyMap["PeerAuthentication"] = 0

	if err := populateMap(argument, policyMap); err != nil {
		return nil, err
	}
	return policyMap, nil
}

func createRuleMap(arguments map[string]string) (map[string]int, error) {
	ruleMap := make(map[string]int)
	// These are the default values
	ruleMap["numPaths"] = 0
	ruleMap["numSourceIP"] = 0
	ruleMap["numNamespaces"] = 0
	ruleMap["numValues"] = 0
	ruleMap["numPrincipals"] = 0

	if err := populateMap(arguments, ruleMap); err != nil {
		return nil, err
	}
	return ruleMap, nil
}

func populateMap(arguments map[string]string, populatedMap map[string]int) error {
	for key := range populatedMap {
		if argVal, inMap := arguments[key]; inMap {
			argVal, err := strconv.Atoi(argVal)
			if err != nil {
				return fmt.Errorf("invalid value: %d", argVal)
			}
			populatedMap[key] = argVal
		}
	}
	return nil
}

func main() {
	securityPtr := flag.String("generate_policy", "numPolicies:1", `List of key value pairs separated by commas.
	Supported options: namespace:string, action:DENY/ALLOW, AuthorizationPolicy:int, PeerAuthentication:in,
	mtlsMode:STRICT/DISABLE, numPrincipals:int, numValues:int, numPaths:int, numSourceIP:int. numNamespaces:int`)

	flag.Parse()

	argumentMap, err := parseArguments(*securityPtr)
	if err != nil {
		fmt.Println(err)
		return
	}

	ruleMap, err := createRuleMap(argumentMap)
	if err != nil {
		fmt.Println(err)
		return
	}

	policyMap, err := createPolicyMap(argumentMap)
	if err != nil {
		fmt.Println(err)
		return
	}

	policiesLeft := 0
	for _, numPolicy := range policyMap {
		policiesLeft += numPolicy
	}
	if policiesLeft <= 0 {
		fmt.Errorf("invalid number of policies: %d", policiesLeft)
	}

	for policyType, numPolicy := range policyMap {
		for i := 1; i <= numPolicy; i++ {
			policy := createPolicyHeader(argumentMap["namespace"], fmt.Sprintf("test-%s-%d", policyType, i), policyType)

			ruleToGenerator := createRuleGeneratorMap(ruleMap)
			if rules, err := generateRules(argumentMap, ruleToGenerator, policy, ruleMap); err != nil {
				fmt.Println(err)
				break
			} else {
				yaml := bytes.Buffer{}
				yaml.WriteString(rules)
				if policiesLeft > 1 {
					yaml.WriteString("---")
				}
				policiesLeft--
				fmt.Println(yaml.String())
			}
		}
	}
}

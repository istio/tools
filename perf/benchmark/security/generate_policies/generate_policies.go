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

	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	authzpb "istio.io/api/security/v1beta1"
)

type ruleGenerator struct {
	gen generator
}

type SecurityPolicy struct {
	AuthZ     AuthorizationPolicy `json:"authZ"`
	Namespace string              `json:"namespace"`
	PeerAuthN PeerAuthentication  `json:"peerAuthN"`
}

type AuthorizationPolicy struct {
	Action        string `json:"action"`
	NumNamespaces int    `json:"numNamespaces"`
	NumPaths      int    `json:"numPaths"`
	NumPolicies   int    `json:"numPolicies"`
	NumPrincipals int    `json:"numPrincipals"`
	NumSourceIP   int    `json:"numSourceIP"`
	NumValues     int    `json:"numValues"`
}

type PeerAuthentication struct {
	MtlsMode    string `json:"mtlsMode"`
	NumPolicies int    `json:"numPolicies"`
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

	createdPolicy, err := ToYAML(spec)
	if err != nil {
		return "", err
	}

	rulesYaml := bytes.Buffer{}
	rulesYaml.WriteString("spec:\n")
	scanner := bufio.NewScanner(strings.NewReader(createdPolicy))
	for scanner.Scan() {
		rulesYaml.WriteString(" " + scanner.Text() + "\n")
	}
	return string(headerYaml) + rulesYaml.String(), nil
}

func createRuleGeneratorMap(authZData AuthorizationPolicy) map[string]*ruleGenerator {
	ruleGeneratorMap := make(map[string]*ruleGenerator)

	if authZData.NumSourceIP > 0 || authZData.NumNamespaces > 0 ||
		authZData.NumPrincipals > 0 {
		ruleGeneratorMap["from"] = &ruleGenerator{
			gen: sourceGenerator{},
		}
	}

	if authZData.NumPaths > 0 {
		ruleGeneratorMap["to"] = &ruleGenerator{
			gen: operationGenerator{},
		}
	}

	if authZData.NumValues > 0 {
		ruleGeneratorMap["when"] = &ruleGenerator{
			gen: conditionGenerator{},
		}
	}
	return ruleGeneratorMap
}

func generateAuthorizationPolicy(policyData SecurityPolicy, policyHeader *MyPolicy) (string, error) {
	spec := &authzpb.AuthorizationPolicy{}
	switch policyData.AuthZ.Action {
	case "ALLOW":
		spec.Action = authzpb.AuthorizationPolicy_ALLOW
	case "DENY":
		spec.Action = authzpb.AuthorizationPolicy_DENY
	case "":
		spec.Action = authzpb.AuthorizationPolicy_DENY
	}

	ruleToGenerator := createRuleGeneratorMap(policyData.AuthZ)
	var ruleList []*authzpb.Rule
	for name := range ruleToGenerator {
		rule := ruleToGenerator[name].gen.generate(policyData)
		ruleList = append(ruleList, rule)
	}
	spec.Rules = ruleList

	yaml, err := PolicyToYAML(policyHeader, spec)
	if err != nil {
		return "", err
	}
	return yaml, nil
}

func generatePeerAuthentication(policyData SecurityPolicy, policyHeader *MyPolicy) (string, error) {
	spec := &authzpb.PeerAuthentication{
		Mtls: &authzpb.PeerAuthentication_MutualTLS{},
	}
	switch policyData.PeerAuthN.MtlsMode {
	case "STRICT":
		spec.Mtls.Mode = authzpb.PeerAuthentication_MutualTLS_STRICT
	case "DISABLE":
		spec.Mtls.Mode = authzpb.PeerAuthentication_MutualTLS_DISABLE
	case "":
		spec.Mtls.Mode = authzpb.PeerAuthentication_MutualTLS_STRICT
	default:
		return "", fmt.Errorf("invalid mtlsMode: %s", policyData.PeerAuthN.MtlsMode)
	}

	yaml, err := PolicyToYAML(policyHeader, spec)
	if err != nil {
		return "", err
	}
	return yaml, nil
}

func generateRules(policyData SecurityPolicy, policyHeader *MyPolicy) (string, error) {
	switch policyHeader.Kind {
	case "AuthorizationPolicy":
		return generateAuthorizationPolicy(policyData, policyHeader)
	case "PeerAuthentication":
		return generatePeerAuthentication(policyData, policyHeader)
	case "RequestAuthentication":
		return "", fmt.Errorf("unimplemented")
	default:
		return "", fmt.Errorf("unknown policy kind: %s", policyHeader.Kind)
	}
}

func createPolicyHeader(namespace string, name string, kind string) *MyPolicy {
	if namespace == "" {
		namespace = "twopods-istio"
	}
	return &MyPolicy{
		APIVersion: "security.istio.io/v1beta1",
		Kind:       kind,
		Metadata:   MetadataStruct{Namespace: namespace, Name: name},
	}
}

func generatePolicy(policyData SecurityPolicy, kind string, numPolicy int, totalPolicies int) (int, error) {
	for i := 1; i <= numPolicy; i++ {
		testName := fmt.Sprintf("test-%s-%d", kind, i)
		policyHeader := createPolicyHeader(policyData.Namespace, testName, kind)

		rules, err := generateRules(policyData, policyHeader)
		if err != nil {
			return 0, err
		}
		yaml := bytes.Buffer{}
		yaml.WriteString(rules)
		if totalPolicies > 1 {
			yaml.WriteString("---")
		}
		totalPolicies--
		fmt.Println(yaml.String())
	}
	return numPolicy, nil
}

func main() {
	configFilePtr := flag.String("configFile", "", "The name of the config json file")
	flag.Parse()

	jsonString := ""
	if *configFilePtr != "" {
		jsonFile, err := os.Open(*configFilePtr)
		if err != nil {
			fmt.Println(err)
		}

		jsonBytes, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			fmt.Println(err)
		}
		jsonString = string(jsonBytes)
	}

	policyData := SecurityPolicy{}
	err := json.Unmarshal([]byte(jsonString), &policyData)
	if err != nil {
		fmt.Println(err)
	}

	policiesLeft := policyData.AuthZ.NumPolicies + policyData.PeerAuthN.NumPolicies
	if policiesLeft <= 0 {
		fmt.Println(fmt.Errorf("invalid number of policies: %d", policiesLeft))
	}

	if policyData.AuthZ.NumPolicies > 0 {
		writtenPolicies, err := generatePolicy(policyData, "AuthorizationPolicy", policyData.AuthZ.NumPolicies, policiesLeft)
		if err != nil {
			fmt.Println(err)
		}
		policiesLeft -= writtenPolicies
	}

	if policyData.PeerAuthN.NumPolicies > 0 {
		_, err := generatePolicy(policyData, "PeerAuthentication", policyData.PeerAuthN.NumPolicies, policiesLeft)
		if err != nil {
			fmt.Println(err)
		}
	}
}

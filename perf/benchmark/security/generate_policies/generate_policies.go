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
	"crypto/rand"
	"crypto/rsa"
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
	AuthZ        AuthorizationPolicy   `json:"authZ"`
	Namespace    string                `json:"namespace"`
	PeerAuthN    PeerAuthentication    `json:"peerAuthN"`
	RequestAuthN RequestAuthentication `json:"requestAuthN"`
}

type AuthorizationPolicy struct {
	Action        string `json:"action"`
	NumNamespaces int    `json:"numNamespaces"`
	NumPaths      int    `json:"numPaths"`
	NumPolicies   int    `json:"numPolicies"`
	NumPrincipals int    `json:"numPrincipals"`
	NumSourceIP   int    `json:"numSourceIP"`
	NumValues     int    `json:"numValues"`
	// The request_principal in the generated authorization policy will match the
	// RequestAuthentication policies generated from the requestAuthN. This allows
	// to test RequestAuthentication and AuthorizationPolicy together to verify that
	// a request with a valid JWT token is allowed.
	NumRequestPrincipals int `json:"numRequestPrincipals"`
}

type PeerAuthentication struct {
	MtlsMode    string `json:"mtlsMode"`
	NumPolicies int    `json:"numPolicies"`
}

type RequestAuthentication struct {
	// Setting InvalidToken to true will create a token which will be signed by it's own
	// privateKey creating a token which will never match with a jwks
	InvalidToken bool   `json:"invalidToken"`
	NumPolicies  int    `json:"numPolicies"`
	NumJwks      int    `json:"numJwks"`
	TokenIssuer  string `json:"tokenIssuer"`
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
		authZData.NumPrincipals > 0 || authZData.NumRequestPrincipals > 0 {
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
	case "DENY", "":
		spec.Action = authzpb.AuthorizationPolicy_DENY
	default:
		return "", fmt.Errorf("action %s not supported", policyData.AuthZ.Action)
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
	case "STRICT", "":
		spec.Mtls.Mode = authzpb.PeerAuthentication_MutualTLS_STRICT
	case "DISABLE":
		spec.Mtls.Mode = authzpb.PeerAuthentication_MutualTLS_DISABLE
	case "PERMISSIVE":
		spec.Mtls.Mode = authzpb.PeerAuthentication_MutualTLS_PERMISSIVE
	default:
		return "", fmt.Errorf("invalid mtlsMode: %s", policyData.PeerAuthN.MtlsMode)
	}

	yaml, err := PolicyToYAML(policyHeader, spec)
	if err != nil {
		return "", err
	}
	return yaml, nil
}

func generateRequestAuthentication(policyData SecurityPolicy, policyHeader *MyPolicy) (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}
	token, err := generateToken(policyData, privateKey)
	if err != nil {
		return "", err
	}
	err = writeTokenIntoFile(token, "token.txt")
	if err != nil {
		return "", err
	}
	jwks, err := generateJwks(privateKey)
	if err != nil {
		return "", err
	}

	var listJWTRules []*authzpb.JWTRule
	if numJwks := policyData.RequestAuthN.NumJwks; numJwks > 0 {
		for i := 1; i <= numJwks; i++ {
			jwkRule := &authzpb.JWTRule{
				Issuer: fmt.Sprintf("issuer-%d", i),
				Jwks:   jwks,
			}
			listJWTRules = append(listJWTRules, jwkRule)
		}
	}

	spec := &authzpb.RequestAuthentication{
		JwtRules: listJWTRules,
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
		return generateRequestAuthentication(policyData, policyHeader)
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

func generatePolicy(policyData SecurityPolicy, kind string, numPolicy int) error {
	for i := 1; i <= numPolicy; i++ {
		testName := fmt.Sprintf("test-%s-%d", strings.ToLower(kind), i)
		policyHeader := createPolicyHeader(policyData.Namespace, testName, kind)

		rules, err := generateRules(policyData, policyHeader)
		if err != nil {
			return err
		}
		yaml := bytes.Buffer{}
		yaml.WriteString(rules)
		yaml.WriteString("---")
		fmt.Println(yaml.String())
	}
	return nil
}

func main() {
	configFilePtr := flag.String("configFile", "", "The name of the config json file")
	flag.Parse()

	jsonBytes := make([]byte, 0)
	if *configFilePtr != "" {
		jsonFile, err := os.Open(*configFilePtr)
		if err != nil {
			fmt.Println(err)
		}

		jsonBytes, err = ioutil.ReadAll(jsonFile)
		if err != nil {
			fmt.Println(err)
		}
	}

	policyData := SecurityPolicy{}
	err := json.Unmarshal(jsonBytes, &policyData)
	if err != nil {
		fmt.Println(err)
	}

	totalPolicies := policyData.AuthZ.NumPolicies + policyData.PeerAuthN.NumPolicies + policyData.RequestAuthN.NumPolicies
	if totalPolicies <= 0 {
		fmt.Println(fmt.Errorf("invalid number of policies: %d", totalPolicies))
	}

	if policyData.AuthZ.NumPolicies > 0 {
		err := generatePolicy(policyData, "AuthorizationPolicy", policyData.AuthZ.NumPolicies)
		if err != nil {
			fmt.Println(err)
		}
	}

	if policyData.PeerAuthN.NumPolicies > 0 {
		err := generatePolicy(policyData, "PeerAuthentication", policyData.PeerAuthN.NumPolicies)
		if err != nil {
			fmt.Println(err)
		}
	}

	if policyData.RequestAuthN.NumPolicies > 0 {
		err := generatePolicy(policyData, "RequestAuthentication", policyData.RequestAuthN.NumPolicies)
		if err != nil {
			fmt.Println(err)
		}
	}
}

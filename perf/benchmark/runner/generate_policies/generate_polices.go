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
    "bytes"
    "github.com/ghodss/yaml"
    "encoding/json"

    authzpb "istio.io/api/security/v1beta1"
)

type ruleOption struct {
    occurance int
    g         generator
}

type MyPolicy struct {
    ApiVersion string `json:"apiVersion"`
    Kind       string `json:"kind"`
    Metadata   MetadataStruct `json:"metadata"`
    Spec       authzpb.AuthorizationPolicy `json:"spec"`
}

type MetadataStruct struct {
    Name      string  `json:"name"`
    Namespace string  `json:"namespace"`
}

func formYaml(policy *MyPolicy) (string, error) {
    header, err := json.Marshal(policy)
    if err != nil {
        return "", err
    }

    yaml, err := yaml.JSONToYAML([]byte(header))
    if err != nil {
        return "", err
    }

    return string(yaml), nil
}

func generateAuthorizationPolicy(action string, ruleToOccurences map[string]*ruleOption, policy *MyPolicy) (string, error) {
    spec := &authzpb.AuthorizationPolicy{}
    // This action will be set by the paramater action
    spec.Action = 1;

    var ruleList []*authzpb.Rule
    for name, ruleOp := range ruleToOccurences {
        rule, err := ruleOp.g.generate(name, ruleOp.occurance)
        if err != nil {
            return "", err
        }
        ruleList = append(ruleList, rule)
    }
    spec.Rules = ruleList

    policy.Spec = *spec

    yaml, err := formYaml(policy)
    if (err != nil) {
        return "", err
    }

    return yaml, nil
}

func generateRule(action string, ruleToOccurences map[string]*ruleOption,
                    policy *MyPolicy) (string, error) {
 
    switch policy.Kind {
    case "AuthorizationPolicy":
        return generateAuthorizationPolicy(action, ruleToOccurences, policy)
    case "PeerAuthentication":
        fmt.Println("PeerAuthentication")
        // TODO implement
        // return generatePeerAuthentication(selector, mtl []*mode, portLevel)
    case "RequestAuthentication":
        fmt.Println("RequestAuthentication")
        // TODO implement
        // return generateRequestAuthentication(selector, ruleToOccurences)
    default:
        fmt.Println("invalid policy")
    }

    return "", fmt.Errorf("invalid policy")
}


func createRules(action string, ruleToOccurences map[string]*ruleOption, policy *MyPolicy) (string, error){
    yaml, err := generateRule(action, ruleToOccurences, policy)
    if (err != nil) {
        return "", err
    }
    return yaml, nil
}

func createPolicyHeader(namespace string, name string, kind string) (*MyPolicy, error) {
    metadata := MetadataStruct{namespace, name}
    
    policy := &MyPolicy{}
    policy.ApiVersion = "security.istio.io/v1beta1"
    policy.Kind = kind
    policy.Metadata = metadata
    return policy, nil
}

func main() {
    yaml := bytes.Buffer{}
    policy, err := createPolicyHeader("deny-method-get", "twopods-istio", "AuthorizationPolicy")
    if (err != nil) {
        fmt.Println(err)
    }

    ruleOptionMap := make(map[string]*ruleOption)
    // These hardcoded values will be provided by 
    // command line arguments passed from runner.py
    ruleOptionMap["when"] = &ruleOption{}
    ruleOptionMap["when"].occurance = 10
    ruleOptionMap["when"].g = conditionGenerator{}

    rules, err := createRules("DENY", ruleOptionMap, policy)
    if (err != nil) {
        fmt.Println(err)
    } else {
        yaml.WriteString(rules)
        fmt.Println(yaml.String())
    }
}

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
)

type ruleOption struct {
    occurance int
    g         generator
}

func generateAuthorizationPolicy(action string, ruleToOccurences map[string]*ruleOption) (string, error) {
    yaml := bytes.Buffer{}
    yaml.WriteString("spec:\n")
    yaml.WriteString(" action: " + action + "\n")

    if (len(ruleToOccurences) == 0) {
        yaml.WriteString(" {}")
    } else {
        yaml.WriteString(" rules:\n")
    }

    for rule, ruleOp := range ruleToOccurences {
        yaml.WriteString("  " + rule + ":\n")
        rules, err := ruleOp.g.generate(rule, ruleOp.occurance)
        if err != nil {
            return "", err
        }
        yaml.WriteString(rules)
    }
    return yaml.String(), nil
}

func generateRule(policy string, action string, ruleToOccurences map[string]*ruleOption) (string, error) {
 
    switch policy {
    case "AuthorizationPolicy":
        return generateAuthorizationPolicy(action, ruleToOccurences)
    case "PeerAuthentication":
        fmt.Println("PeerAuthentication")
        // TODO impliment
        // return generatePeerAuthentication(selector, mtl []*mode, portLevel)
    case "RequestAuthentication":
        fmt.Println("RequestAuthentication")
        // TODO impliment
        // return generateRequestAuthentication(selector, ruleToOccurences)
    default:
        fmt.Println("invalid policy")
    }

    return "", fmt.Errorf("invalid policy")
}


func createRules(policy string, action string, ruleToOccurences map[string]*ruleOption) (string, error){
    yaml, err := generateRule(policy, action, ruleToOccurences)
    if (err != nil) {
        return "", err
    }
    return yaml, nil
}

func createHeader(namespace string, name string, kind string) (string, error) {
    // This could be updated to exporting the data into a file,
    // instead of printing it
    yaml := bytes.Buffer{}
    header := `apiVersion: security.istio.io/v1beta1
kind: %s
metadata:
  name: %s
  namespace: %s
`
    yaml.WriteString(fmt.Sprintf(header, kind, name, namespace))
    return yaml.String(), nil
}

func main() {
    yaml := bytes.Buffer{}
    header, err := createHeader("foo", "httpbin", "AuthorizationPolicy")
    if (err != nil) {
        fmt.Println(err)
    } else {
        yaml.WriteString(header)
    }

    ruleOptionMap := make(map[string]*ruleOption)
    // These hardcoded values will be provided by 
    // command line arguments passed from runner.py
    ruleOptionMap["from"] = &ruleOption{}
    ruleOptionMap["from"].occurance = 2
    ruleOptionMap["from"].g = sourceGenerator{}

    ruleOptionMap["when"] = &ruleOption{}
    ruleOptionMap["when"].occurance = 1
    ruleOptionMap["when"].g = conditionGenerator{}

    rules, err := createRules("AuthorizationPolicy", "DENY", ruleOptionMap)
    if (err != nil) {
        fmt.Println(err)
    } else {
        yaml.WriteString(rules)
        fmt.Println(yaml.String())
    }
}

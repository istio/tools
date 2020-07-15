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
    "flag"
    "bytes"
    "strings"
    "bufio"
    "sort"
    "github.com/ghodss/yaml"
    "encoding/json"
    "istio.io/istio/pkg/util/protomarshal"

    authzpb "istio.io/api/security/v1beta1"
)

type ruleOption struct {
    occurence int
    gen         generator
}

type MyPolicy struct {
    ApiVersion string `json:"apiVersion"`
    Kind       string `json:"kind"`
    Metadata   MetadataStruct `json:"metadata"`
}

type MetadataStruct struct {
    Name      string  `json:"name"`
    Namespace string  `json:"namespace"`
}

func ToYAML(policy *MyPolicy, spec *authzpb.AuthorizationPolicy) (string, error) {
    header, err := json.Marshal(policy)
    if err != nil {
        return "", err
    }

    headerYaml, err := yaml.JSONToYAML([]byte(header))
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

func getOrderedKeySlice(ruleToOccurences map[string]*ruleOption) *[]string {
    var sortedKeys []string
    for key, _ := range ruleToOccurences {
        sortedKeys = append(sortedKeys, key)
    }
    sort.Sort(sort.StringSlice(sortedKeys))
    return &sortedKeys
}

func generateAuthorizationPolicy(action string, ruleToOccurences map[string]*ruleOption, policy *MyPolicy) (string, error) {
    spec := &authzpb.AuthorizationPolicy{}
    switch action {
    case "ALLOW":
        spec.Action = authzpb.AuthorizationPolicy_ALLOW
    case "DENY":
        spec.Action = authzpb.AuthorizationPolicy_DENY
    }

    var ruleList []*authzpb.Rule
    sortedKeys := getOrderedKeySlice(ruleToOccurences)
    for _, name := range *sortedKeys {
        ruleOp := ruleToOccurences[name] 
        rule, err := ruleOp.gen.generate(name, ruleOp.occurence, action)
        if err != nil {
            return "", err
        }
        ruleList = append(ruleList, rule)
    }
    spec.Rules = ruleList

    yaml, err := ToYAML(policy, spec)
    if err != nil {
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
    if err != nil {
        return "", err
    }
    return yaml, nil
}

func createPolicyHeader(namespace string, name string, kind string) (*MyPolicy, error) {
    metadata := MetadataStruct{namespace, name}
    return &MyPolicy{
        ApiVersion: "security.istio.io/v1beta1",
        Kind: kind,
        Metadata: metadata,
      }, nil
}

func createRuleOptionMap(ruleToOccurancesPtr map[string]*int) *map[string]*ruleOption {
    ruleOptionMap := make(map[string]*ruleOption)
    for rule, occurence := range ruleToOccurancesPtr {
        ruleOptionMap[rule] = &ruleOption{}
        ruleOptionMap[rule].occurence = *occurence
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
    return &ruleOptionMap;
}


func main() {
    namespacePtr := flag.String("namespace", "twopods-istio", "Current namespace")
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
        policy, err := createPolicyHeader(fmt.Sprintf("%s%d", "test-", i), *namespacePtr, *policyType)
        if err != nil {
            fmt.Println(err)
        }

        ruleOptionMap := createRuleOptionMap(ruleToOccurancesPtr)
        rules, err := createRules(*actionPtr, *ruleOptionMap, policy)
        if err != nil {
            fmt.Println(err)
        } else {
            yaml.WriteString(rules)
            fmt.Println(yaml.String())
        }
    }
}

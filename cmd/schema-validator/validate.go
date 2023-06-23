// Copyright 2020 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"sigs.k8s.io/yaml"

	"istio.io/tools/pkg/schemavalidation"
)

func main() {
	var documentPath, schemaPath string
	flag.StringVar(&documentPath, "documentPath", "", "path to document to validate (YAML/JSON)")
	flag.StringVar(&schemaPath, "schemaPath", "", "path to schema to validate against (JSON)")
	flag.Parse()

	fmt.Printf("Validating %s against %s\n", documentPath, schemaPath)
	doc, err := os.ReadFile(documentPath)
	if err != nil {
		log.Fatalf("fail to read: %v", err)
	}
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("fail to read: %v", err)
	}

	if err := schemavalidation.Validate(doc, schema); err != nil {
		log.Fatal(err)
	}
	log.Println("document is valid")
}

func readYAMLAsJSON(filename string) (string, error) {
	documentContents, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %s", err.Error())
	}

	documentAsYaml, err := yaml.YAMLToJSON(documentContents)
	if err != nil {
		return "", fmt.Errorf("unable to parse YAML: %s", err.Error())
	}

	return string(documentAsYaml), nil
}

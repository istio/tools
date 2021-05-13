// Copyright 2020 Istio Authors
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
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

func main() {
	var documentPath, schemaPath string
	flag.StringVar(&documentPath, "documentPath", "", "path to document to validate (YAML/JSON)")
	flag.StringVar(&schemaPath, "schemaPath", "", "path to schema to validate against (JSON)")
	flag.Parse()

	fmt.Printf("Validating %s against %s\n", documentPath, schemaPath)
	schemaLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", schemaPath))

	documentContents, err := readYAMLAsJSON(documentPath)
	if err != nil {
		fmt.Printf("Could not parse document: %s\n", err.Error())
		os.Exit(1)
	}
	documentLoader := gojsonschema.NewStringLoader(documentContents)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		fmt.Printf("Unable to validate: %s\n", err.Error())
		os.Exit(2)
	}

	if result.Valid() {
		fmt.Printf("The document is valid\n")
	} else {
		fmt.Printf("The document is not valid. See errors :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		os.Exit(3)
	}
}

func readYAMLAsJSON(filename string) (string, error) {
	documentContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %s", err.Error())
	}

	documentAsYaml, err := yaml.YAMLToJSON(documentContents)
	if err != nil {
		return "", fmt.Errorf("unable to parse YAML: %s", err.Error())
	}

	return string(documentAsYaml), nil
}

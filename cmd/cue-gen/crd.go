// Copyright 2019 Istio Authors
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
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"cuelang.org/go/encoding/openapi"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
)

var statusOutput = `
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: null
  storedVersions: null`

// Build CRDs based on the configuration and schema.
func completeCRD(c *apiextv1beta1.CustomResourceDefinition, versionSchemas map[string]*openapi.OrderedMap) {

	for i, version := range c.Spec.Versions {

		b, err := versionSchemas[version.Name].MarshalJSON()
		if err != nil {
			log.Fatalf("Cannot marshal OpenAPI schema for %v: %v", c.Name, err)
		}

		j := &apiextv1beta1.JSONSchemaProps{}
		if err = json.Unmarshal(b, j); err != nil {
			log.Fatalf("Cannot unmarshal raw OpenAPI schema to JSONSchemaProps for %v: %v", c.Name, err)
		}

		version.Schema = &apiextv1beta1.CustomResourceValidation{OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]apiextv1beta1.JSONSchemaProps{
				"spec": *j,
			},
		}}

		fmt.Printf("Checking if the schema is structural for %v \n", c.Name)
		if err = validateStructural(version.Schema.OpenAPIV3Schema); err != nil {
			log.Fatal(err)
		}

		c.Spec.Versions[i] = version
	}

	if schemasEqual(versionSchemas) {
		collapseCRDVersions(c)
	}

	c.APIVersion = apiextv1beta1.SchemeGroupVersion.String()
	c.Kind = "CustomResourceDefinition"

	// marshal to an empty field in the output
	c.Status = apiextv1beta1.CustomResourceDefinitionStatus{}
}

func validateStructural(s *apiextv1beta1.JSONSchemaProps) error {
	out := &apiext.JSONSchemaProps{}
	if err := apiextv1beta1.Convert_v1beta1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(s, out, nil); err != nil {
		return fmt.Errorf("cannot convert v1beta1 JSONSchemaProps to v1 JSONSchemaProps: %v", err)
	}

	r, err := structuralschema.NewStructural(out)
	if err != nil {
		return fmt.Errorf("cannot convert to a structural schema: %v", err)
	}

	if errs := structuralschema.ValidateStructural(r, nil); len(errs) != 0 {
		return fmt.Errorf("schema is not structural: %v", errs.ToAggregate().Error())
	}

	return nil
}

func schemasEqual(versionSchemas map[string]*openapi.OrderedMap) bool {
	if len(versionSchemas) < 2 {
		return true
	}
	var schema *openapi.OrderedMap
	for _, s := range versionSchemas {
		if schema == nil {
			schema = s
			continue
		}
		if !reflect.DeepEqual(*schema, *s) {
			return false
		}
	}
	return true
}

func collapseCRDVersions(c *apiextv1beta1.CustomResourceDefinition) {
	c.Spec.Validation = c.Spec.Versions[0].Schema
	c.Spec.AdditionalPrinterColumns = c.Spec.Versions[0].AdditionalPrinterColumns
	for i := range c.Spec.Versions {
		c.Spec.Versions[i].Schema = nil
		c.Spec.Versions[i].AdditionalPrinterColumns = []apiextv1beta1.CustomResourceColumnDefinition{}
	}
}

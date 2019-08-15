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

	"cuelang.org/go/encoding/openapi"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Interim solution to build the Istio CRDs before we move to KubeBuilder.
func (x *builder) getCRD(crdCfg CrdConfig, schema interface{}) apiextv1beta1.CustomResourceDefinition {
	// boilerplate OrderMap for CRD spec
	m := &openapi.OrderedMap{}
	m.Set("spec", schema)
	kvs := []openapi.KeyValue{
		{
			Key:   "type",
			Value: "object",
		},
		{
			Key:   "properties",
			Value: m,
		},
	}
	schemaMap := &openapi.OrderedMap{}
	schemaMap.SetAll(kvs)

	// convert the schema from an OrderedMap to JSONSchemaProps
	b, err := schemaMap.MarshalJSON()
	if err != nil {
		log.Fatalf("Cannot marshal OpenAPI schema for %v", crdCfg.Metadata.Name)
	}
	j := &apiextv1beta1.JSONSchemaProps{}
	if err = json.Unmarshal(b, j); err != nil {
		log.Fatalf("Cannot unmarshal raw OpenAPI schema to JSONSchemaProps for %v", crdCfg.Metadata.Name)
	}

	fmt.Printf("Checking if the schema is structural for %v \n", crdCfg.Metadata.Name)
	if err = validateStructural(j); err != nil {
		log.Fatal(err)
	}

	crdCfg.Spec.Validation = &apiextv1beta1.CustomResourceValidation{OpenAPIV3Schema: j}

	crd := apiextv1beta1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiextv1beta1.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: crdCfg.Metadata,
		Spec:       crdCfg.Spec,
	}

	// marshal to an empty field in the output
	crd.Status.Conditions = []apiextv1beta1.CustomResourceDefinitionCondition{}
	crd.Status.StoredVersions = []string{}

	return crd
}

func validateStructural(s *apiextv1beta1.JSONSchemaProps) error {
	out := &apiext.JSONSchemaProps{}
	if err := apiextv1beta1.Convert_v1beta1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(s, out, nil); err != nil {
		return fmt.Errorf("Cannot convert v1beta1 JSONSchemaProps to v1 JSONSchemaProps: %v", err)
	}

	r, err := structuralschema.NewStructural(out)
	if err != nil {
		return fmt.Errorf("Cannot convert to a structural schema: %v", err)
	}

	if errs := structuralschema.ValidateStructural(r, nil); len(errs) != 0 {
		return fmt.Errorf("Schema is not structural: %v", errs.ToAggregate().Error())
	}

	return nil
}

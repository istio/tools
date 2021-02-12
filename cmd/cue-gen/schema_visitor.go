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
	"strings"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/pointer"
	crdutil "sigs.k8s.io/controller-tools/pkg/crd"
)

var _ crdutil.SchemaVisitor = &preserveUnknownFieldVisitor{}

// a visitor to add x-kubernetes-preserve-unknown-field to a schema
type preserveUnknownFieldVisitor struct {
	// path is in the format of a.b.c to indicate a field path in the schema
	// a `[]` indicates an array, a string is a key to a map in the schema
	// e.g. a.[].b.c
	path string
}

func (v *preserveUnknownFieldVisitor) Visit(schema *apiextv1.JSONSchemaProps) crdutil.SchemaVisitor {
	if schema == nil {
		return v
	}
	p := strings.Split(v.path, ".")
	if len(p) == 0 {
		return nil
	}
	if len(p) == 1 {
		if s, ok := schema.Properties[p[0]]; ok {
			s.XPreserveUnknownFields = pointer.BoolPtr(true)
			schema.Properties[p[0]] = s
		}
		return nil
	}
	if len(p) > 1 {
		if p[0] == "[]" && schema.Items == nil {
			return nil
		}
		if p[0] != "[]" && schema.Items != nil {
			return nil
		}
		if _, ok := schema.Properties[p[0]]; p[0] != "[]" && !ok {
			return nil
		}
		return &preserveUnknownFieldVisitor{path: strings.Join(p[1:], ".")}
	}
	return nil
}

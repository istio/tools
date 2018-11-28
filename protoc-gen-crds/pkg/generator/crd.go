// Copyright 2018 Istio Authors
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

package generator

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"istio.io/tools/kubernetes/resource"
	"istio.io/tools/protoc-gen-crds/pkg/crd"
	"istio.io/tools/protoc-gen-crds/pkg/crd/openapi"
	"istio.io/tools/protoc-gen-crds/pkg/naming"
	"istio.io/tools/protoc-gen-crds/pkg/pmodel"
)

type crdBuilder struct {
	model *pmodel.Model

	// The files that were passed as input to the generator
	filesToGenerate map[string]struct{}

	// Preallocated mapping of descriptors to ObjectTypes
	descriptorsToTypes map[*descriptor.DescriptorProto]*openapi.Object

	definitions []*crd.ResourceDefinition
}

func newCRDBuilder(model *pmodel.Model) *crdBuilder {
	return &crdBuilder{
		model:              model,
		filesToGenerate:    make(map[string]struct{}),
		descriptorsToTypes: make(map[*descriptor.DescriptorProto]*openapi.Object),
	}
}

func (b *crdBuilder) build() error {

	// Create a lookup table for names
	for _, f := range b.model.Request.GetFileToGenerate() {
		b.filesToGenerate[f] = struct{}{}
	}

	for _, pf := range b.model.Request.ProtoFile {
		if !b.isFileToGenerate(pf.Name) {
			continue
		}

		// Find and extract specs
		for _, msg := range pf.MessageType {
			if err := b.generateDefinition(pf, msg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *crdBuilder) isFileToGenerate(name *string) bool {
	if name == nil {
		return false
	}

	_, ok := b.filesToGenerate[*name]
	return ok
}

func (b *crdBuilder) generateDefinition(pf *descriptor.FileDescriptorProto, m *descriptor.DescriptorProto) error {

	defn, err := b.readOptions(pf, m)
	if err != nil {
		return err
	}

	if defn == nil {
		return nil
	}

	t, err := b.getObjectType(m)
	if err != nil {
		return err
	}
	t.Name = naming.PascalCase(defn.Spec.Names.Kind) + "Spec"
	defn.Spec.Validation = t

	b.definitions = append(b.definitions, defn)

	return nil
}

func (b *crdBuilder) getObjectType(m *descriptor.DescriptorProto) (*openapi.Object, error) {

	t := b.descriptorsToTypes[m]
	if t != nil {
		return t, nil
	}

	t = &openapi.Object{
		Name: *m.Name,
	}
	b.descriptorsToTypes[m] = t

	for _, field := range m.Field {
		f := &openapi.Field{
			Name: *field.Name,
		}

		var ft openapi.Type
		switch *field.Type {
		case descriptor.FieldDescriptorProto_TYPE_INT32:
			ft = openapi.NewInt32()
		case descriptor.FieldDescriptorProto_TYPE_STRING:
			ft = openapi.NewString()
		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			ft = openapi.NewBool()
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			targetDescriptor := b.model.FindType(*field.TypeName)
			var err error
			if ft, err = b.getObjectType(targetDescriptor); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unrecognized type: %v", field.Type)
		}

		if field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
			ft = &openapi.Array{ElementType: ft}
		}
		f.Type = ft

		t.Fields = append(t.Fields, f)
	}

	b.descriptorsToTypes[m] = t

	return t, nil
}

func (b *crdBuilder) readOptions(pf *descriptor.FileDescriptorProto, m *descriptor.DescriptorProto) (*crd.ResourceDefinition, error) {
	d := &crd.ResourceDefinition{
		APIVersion: "apiextensions.k8s.io/v1beta1",
		Kind:       "CustomResourceDefinition",
	}

	if m.Options == nil {
		return nil, nil
	}

	isSpec, err := pmodel.GetOptionBool(m.Options, resource.E_Spec)
	if err != nil {
		return nil, err
	}

	if !isSpec {
		return nil, nil
	}

	scope, err := pmodel.GetOptionScope(m.Options, resource.E_Scope)
	if err != nil {
		return nil, err
	}
	switch scope {
	case resource.Scope_NAMESPACED:
		d.Spec.Scope = crd.Namespaced

	case resource.Scope_CLUSTER:
		d.Spec.Scope = crd.Cluster
	default:
		return nil, fmt.Errorf("unrecognized scope encountered: %v", scope)
	}

	if d.Spec.Group, err = pmodel.GetOptionString(m.Options, resource.E_Group); err != nil {
		return nil, err
	}
	if d.Spec.Version, err = pmodel.GetOptionString(m.Options, resource.E_Version); err != nil {
		return nil, err
	}
	if d.Spec.Names.Kind, err = pmodel.GetOptionString(m.Options, resource.E_Kind); err != nil {
		return nil, err
	}
	if d.Spec.Names.Singular, err = pmodel.GetOptionString(m.Options, resource.E_Singular); err != nil {
		return nil, err
	}
	if d.Spec.Names.Plural, err = pmodel.GetOptionString(m.Options, resource.E_Plural); err != nil {
		return nil, err
	}

	// Apply defaults

	if d.Spec.Group == "" {
		// Default to the package Name
		d.Spec.Group = *pf.Package
	}

	if d.Spec.Version == "" {
		// Default to v1
		d.Spec.Version = "v1"
	}

	if d.Spec.Names.Kind == "" {
		// Default to message Name
		d.Spec.Names.Kind = *m.Name
	}

	if d.Spec.Names.Singular == "" {
		// Default to message Name
		d.Spec.Names.Singular = strings.ToLower(d.Spec.Names.Kind)
	}

	if d.Spec.Names.Plural == "" {
		// Default to Singular + "s"
		d.Spec.Names.Plural = d.Spec.Names.Singular + "s"
	}

	d.Metadata.Name = d.Spec.Names.Plural + "." + d.Spec.Group

	return d, nil
}

//
//func (c *crdBuilder) generateResponse() error {
//
//	for _, defn := range c.model.Definitions {
//		f := &plugin.CodeGeneratorResponse_File{}
//		p := getOutputPath(defn)
//		f.Name = &p
//
//		content, err := golang.Emit(defn)
//		if err != nil {
//			return err
//		}
//
//		f.Content = &content
//
//		c.response.File = append(c.response.File, f)
//	}
//
//	contents, err := crd.EmitCustomResourceDefinitions(c.model)
//	if err != nil {
//		return err
//	}
//
//	if contents != "" {
//		f := &plugin.CodeGeneratorResponse_File{}
//		name := "crd.yaml"
//		f.Name = &name
//		f.Content = &contents
//
//		c.response.File = append(c.response.File, f)
//	}
//
//	return nil
//}
//
//func errorResponse(err error) *plugin.CodeGeneratorResponse {
//	e := err.Error()
//	return &plugin.CodeGeneratorResponse{
//		Error: &e,
//	}
//}
//

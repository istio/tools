// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
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
	"bytes"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"

	"istio.io/tools/pkg/protomodel"
)

type openapiGenerator struct {
	buffer bytes.Buffer
	model  *protomodel.Model
	mode   bool

	// transient state as individual files are processed
	currentPackage             *protomodel.PackageDescriptor
	currentFrontMatterProvider *protomodel.FileDescriptor

	schemas map[*protomodel.MessageDescriptor]*openapi3.Schema
}

func newOpenAPIGenerator(model *protomodel.Model, mode bool) *openapiGenerator {
	return &openapiGenerator{
		model:   model,
		mode:    mode,
		schemas: make(map[*protomodel.MessageDescriptor]*openapi3.Schema),
	}
}

func (g *openapiGenerator) generateOutput(filesToGen map[*protomodel.FileDescriptor]bool) (*plugin.CodeGeneratorResponse, error) {
	// process each package; we produce one output file per package
	response := plugin.CodeGeneratorResponse{}

	for _, pkg := range g.model.Packages {
		g.currentPackage = pkg
		g.currentFrontMatterProvider = pkg.FileDesc()

		// anything to output for this package?
		count := 0
		for _, file := range pkg.Files {
			if _, ok := filesToGen[file]; ok {
				count++
			}
		}

		if count > 0 {
			g.generatePerPackageOutput(filesToGen, pkg, &response)
		}
	}

	return &response, nil
}

func (g *openapiGenerator) getFileContents(file *protomodel.FileDescriptor,
	messages map[string]*protomodel.MessageDescriptor,
	enums map[string]*protomodel.EnumDescriptor,
	services map[string]*protomodel.ServiceDescriptor) {
	for _, m := range file.AllMessages {
		messages[g.relativeName(m)] = m
	}

	for _, e := range file.AllEnums {
		enums[g.relativeName(e)] = e
	}

	for _, s := range file.Services {
		services[g.relativeName(s)] = s
	}
}

func (g *openapiGenerator) generatePerPackageOutput(filesToGen map[*protomodel.FileDescriptor]bool, pkg *protomodel.PackageDescriptor,
	response *plugin.CodeGeneratorResponse) {
	// We need to produce a file for this package.

	// Decide which types need to be included in the generated file.
	// This will be all the types in the fileToGen input files, along with any
	// dependent types which are located in packages that don't have
	// a known location on the web.
	messages := make(map[string]*protomodel.MessageDescriptor)
	enums := make(map[string]*protomodel.EnumDescriptor)
	services := make(map[string]*protomodel.ServiceDescriptor)

	for _, file := range pkg.Files {
		if _, ok := filesToGen[file]; ok {
			g.getFileContents(file, messages, enums, services)
		}
	}

	rf := g.generateFile(pkg.Name, pkg.FileDesc(), messages, enums, services)
	response.File = append(response.File, &rf)
}

type allSchemas struct {
	Schemas []*openapi3.Schema
}

// Generate an OpenAPI spec for a collection of cross-linked files.
func (g *openapiGenerator) generateFile(name string,
	_ *protomodel.FileDescriptor,
	messages map[string]*protomodel.MessageDescriptor,
	_ map[string]*protomodel.EnumDescriptor,
	_ map[string]*protomodel.ServiceDescriptor) plugin.CodeGeneratorResponse_File {

	for _, message := range messages {
		s := openapi3.NewSchema()
		g.schemas[message] = s

		for _, field := range message.Fields {
			s.WithProperty(field.GetName(), g.fieldType(field))
		}
	}

	var schemas allSchemas
	for _, schema := range g.schemas {
		schemas.Schemas = append(schemas.Schemas, schema)
	}

	g.buffer.Reset()
	b, _ := yaml.Marshal(&schemas)
	g.buffer.Write(b)

	return plugin.CodeGeneratorResponse_File{
		Name:    proto.String(name + ".yaml"),
		Content: proto.String(g.buffer.String()),
	}
}

func (g *openapiGenerator) fieldType(field *protomodel.FieldDescriptor) *openapi3.Schema {
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_FLOAT, descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return openapi3.NewFloat64Schema()

	case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		return openapi3.NewInt32Schema()

	case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return openapi3.NewInt64Schema()

	case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
		return openapi3.NewInt64Schema()

	case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
		return openapi3.NewInt32Schema()

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return openapi3.NewBoolSchema()

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		return openapi3.NewStringSchema()

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		/*
			msg := field.FieldType.(*protomodel.MessageDescriptor)
			if msg.GetOptions().GetMapEntry() {
				keyType := g.fieldTypeName(msg.Fields[0])
				valType := g.linkify(msg.Fields[1].FieldType, g.fieldTypeName(msg.Fields[1]))
				return "map&lt;" + keyType + ",&nbsp;" + valType + "&gt;"
			}
			name = g.relativeName(field.FieldType)
		*/
		return nil

	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		return openapi3.NewBytesSchema()

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		return openapi3.NewInt64Schema()
	}

	return nil
	/*
		if field.IsRepeated() {
			name += "[]"
		}

		if field.OneofIndex != nil {
			name += " (oneof)"
		}

		return name
	*/
}

func (g *openapiGenerator) relativeName(desc protomodel.CoreDesc) string {
	typeName := protomodel.DottedName(desc)
	if desc.PackageDesc() == g.currentPackage {
		return typeName
	}

	return desc.PackageDesc().Name + "." + typeName
}

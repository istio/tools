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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
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
}

func newOpenAPIGenerator(model *protomodel.Model, mode bool) *openapiGenerator {
	return &openapiGenerator{
		model: model,
		mode:  mode,
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

// Generate an OpenAPI spec for a collection of cross-linked files.
func (g *openapiGenerator) generateFile(name string,
	_ *protomodel.FileDescriptor,
	messages map[string]*protomodel.MessageDescriptor,
	enums map[string]*protomodel.EnumDescriptor,
	_ map[string]*protomodel.ServiceDescriptor) plugin.CodeGeneratorResponse_File {

	allSchemas := make(map[string]*openapi3.SchemaRef)

	for _, message := range messages {
		if message.Parent == nil {
			g.generateMessage(message, allSchemas)
		}
	}

	for _, enum := range enums {
		// when there is no parent to the enum.
		if len(enum.QualifiedName()) == 1 {
			g.generateEnum(enum, allSchemas)
		}
	}

	c := openapi3.NewComponents()
	c.Schemas = allSchemas

	g.buffer.Reset()
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to marshall the output of %v to json", name)
	}
	g.buffer.Write(b)

	return plugin.CodeGeneratorResponse_File{
		Name:    proto.String(name + ".json"),
		Content: proto.String(g.buffer.String()),
	}
}

func (g *openapiGenerator) generateMessage(message *protomodel.MessageDescriptor, allSchemas map[string]*openapi3.SchemaRef) {
	if o := g.generateMessageSchema(message); o != nil {
		allSchemas[g.absoluteName(message)] = o.NewRef()
	}
}

func (g *openapiGenerator) generateMessageSchema(message *protomodel.MessageDescriptor) *openapi3.Schema {
	// skip MapEntry message because we handle map using the map's repeated field.
	if message.GetOptions().GetMapEntry() {
		return nil
	}
	o := openapi3.NewObjectSchema()
	o.Description = g.generateDescription(message)
	oneof := make(map[int32][]*openapi3.Schema)
	for _, field := range message.Fields {
		// skip deprecated field as it is not supported by Kubernetes yet.
		if field.GetOptions().GetDeprecated() {
			continue
		}
		if field.OneofIndex == nil {
			o.WithProperty(field.GetName(), g.fieldType(field))
		} else {
			oneof[*field.OneofIndex] = append(oneof[*field.OneofIndex], g.fieldType(field))
		}
	}
	for k, v := range oneof {
		o.WithProperty(message.GetOneofDecl()[k].GetName(), openapi3.NewOneOfSchema(v...))
	}
	return o
}

func (g *openapiGenerator) generateEnum(enum *protomodel.EnumDescriptor, allSchemas map[string]*openapi3.SchemaRef) {
	o := g.generateEnumSchema(enum)
	allSchemas[g.absoluteName(enum)] = o.NewRef()
}

func (g *openapiGenerator) generateEnumSchema(enum *protomodel.EnumDescriptor) *openapi3.Schema {
	o := openapi3.NewStringSchema()
	o.Description = g.generateDescription(enum)
	values := enum.GetValue()
	enumNames := make([]string, len(values))
	for i, v := range values {
		enumNames[i] = v.GetName()
	}
	o.WithEnum(enumNames)
	return o
}

func (g *openapiGenerator) absoluteName(desc protomodel.CoreDesc) string {
	typeName := protomodel.DottedName(desc)
	return desc.PackageDesc().Name + "." + typeName
}

// converts the first section of the leading comment or the description of the proto
// to a single line of description.
func (g *openapiGenerator) generateDescription(desc protomodel.CoreDesc) string {
	c := strings.TrimSpace(desc.Location().GetLeadingComments())
	t := strings.Split(c, "\n\n")[0]
	// omit the comment that starts with `$`.
	if strings.HasPrefix(t, "$") {
		return ""
	}
	return strings.Join(strings.Fields(t), " ")
}

func (g *openapiGenerator) fieldType(field *protomodel.FieldDescriptor) *openapi3.Schema {
	var schema *openapi3.Schema
	var isMap bool
	switch *field.Type {
	case descriptor.FieldDescriptorProto_TYPE_FLOAT, descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		schema = openapi3.NewFloat64Schema()

	case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		schema = openapi3.NewInt32Schema()

	case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		schema = openapi3.NewInt64Schema()

	case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
		schema = openapi3.NewInt64Schema()

	case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
		schema = openapi3.NewInt32Schema()

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		schema = openapi3.NewBoolSchema()

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		schema = openapi3.NewStringSchema()

	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		msg := field.FieldType.(*protomodel.MessageDescriptor)
		if msg.GetOptions().GetMapEntry() {
			isMap = true
			schema = openapi3.NewObjectSchema().WithAdditionalProperties(g.fieldType(msg.Fields[1]))
		} else {
			schema = g.generateMessageSchema(msg)
		}

	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		schema = openapi3.NewBytesSchema()

	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		enum := field.FieldType.(*protomodel.EnumDescriptor)
		schema = g.generateEnumSchema(enum)
	}

	if field.IsRepeated() && !isMap {
		schema = openapi3.NewArraySchema().WithItems(schema)
	}

	if schema != nil {
		schema.Description = g.generateDescription(field)
	}

	return schema
}

func (g *openapiGenerator) relativeName(desc protomodel.CoreDesc) string {
	typeName := protomodel.DottedName(desc)
	if desc.PackageDesc() == g.currentPackage {
		return typeName
	}

	return desc.PackageDesc().Name + "." + typeName
}

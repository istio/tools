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
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"

	"istio.io/tools/pkg/protomodel"
)

// Some special types with predefined schemas.
var specialTypes = map[string]openapi3.Schema{
	"google.protobuf.Struct": {
		Properties: map[string]*openapi3.SchemaRef{
			"fields": {
				Value: openapi3.NewObjectSchema().WithAnyAdditionalProperties(),
			},
		},
	},
	"google.protobuf.ListValue": {
		Properties: map[string]*openapi3.SchemaRef{
			"values": {
				Value: openapi3.NewArraySchema().WithItems(openapi3.NewObjectSchema()),
			},
		},
	},
}

type openapiGenerator struct {
	buffer     bytes.Buffer
	model      *protomodel.Model
	perFile    bool
	singleFile bool
	yaml       bool
	useRef     bool

	// transient state as individual files are processed
	currentPackage             *protomodel.PackageDescriptor
	currentFrontMatterProvider *protomodel.FileDescriptor

	messages map[string]*protomodel.MessageDescriptor
}

func newOpenAPIGenerator(model *protomodel.Model, perFile bool, singleFile bool, yaml bool, useRef bool) *openapiGenerator {
	return &openapiGenerator{
		model:      model,
		perFile:    perFile,
		singleFile: singleFile,
		yaml:       yaml,
		useRef:     useRef,
	}
}

func (g *openapiGenerator) generateOutput(filesToGen map[*protomodel.FileDescriptor]bool) (*plugin.CodeGeneratorResponse, error) {
	response := plugin.CodeGeneratorResponse{}

	if g.singleFile {
		g.generateSingleFileOutput(filesToGen, &response)
	} else {
		for _, pkg := range g.model.Packages {
			g.currentPackage = pkg

			// anything to output for this package?
			count := 0
			for _, file := range pkg.Files {
				if _, ok := filesToGen[file]; ok {
					count++
				}
			}

			if count > 0 {
				if g.perFile {
					g.generatePerFileOutput(filesToGen, pkg, &response)
				} else {
					g.generatePerPackageOutput(filesToGen, pkg, &response)
				}
			}
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

func (g *openapiGenerator) generatePerFileOutput(filesToGen map[*protomodel.FileDescriptor]bool, pkg *protomodel.PackageDescriptor,
	response *plugin.CodeGeneratorResponse) {

	for _, file := range pkg.Files {
		if _, ok := filesToGen[file]; ok {
			g.currentFrontMatterProvider = file
			messages := make(map[string]*protomodel.MessageDescriptor)
			enums := make(map[string]*protomodel.EnumDescriptor)
			services := make(map[string]*protomodel.ServiceDescriptor)

			g.getFileContents(file, messages, enums, services)
			filename := path.Base(file.GetName())
			extension := path.Ext(filename)
			name := filename[0 : len(filename)-len(extension)]

			rf := g.generateFile(name, file, messages, enums, services)
			response.File = append(response.File, &rf)
		}
	}

}

func (g *openapiGenerator) generateSingleFileOutput(filesToGen map[*protomodel.FileDescriptor]bool, response *plugin.CodeGeneratorResponse) {
	messages := make(map[string]*protomodel.MessageDescriptor)
	enums := make(map[string]*protomodel.EnumDescriptor)
	services := make(map[string]*protomodel.ServiceDescriptor)

	for file, ok := range filesToGen {
		if ok {
			g.getFileContents(file, messages, enums, services)
		}
	}

	rf := g.generateFile("openapiv3", &protomodel.FileDescriptor{}, messages, enums, services)
	response.File = []*plugin.CodeGeneratorResponse_File{&rf}
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

	g.currentFrontMatterProvider = pkg.FileDesc()

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
	pkg *protomodel.FileDescriptor,
	messages map[string]*protomodel.MessageDescriptor,
	enums map[string]*protomodel.EnumDescriptor,
	_ map[string]*protomodel.ServiceDescriptor) plugin.CodeGeneratorResponse_File {

	g.messages = messages

	allSchemas := make(map[string]*openapi3.SchemaRef)

	for _, message := range messages {
		// we generate the top-level messages here and the nested messages are generated
		// inside each top-level message.
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

	var version string
	var description string
	// only get the API version when generate per package or per file,
	// as we cannot guarantee all protos in the input are the same version.
	if !g.singleFile {
		if g.currentFrontMatterProvider != nil && g.currentFrontMatterProvider.Matter.Description != "" {
			description = g.currentFrontMatterProvider.Matter.Description
		} else if pd := g.generateDescription(g.currentPackage); pd != "" {
			description = pd
		} else {
			description = "OpenAPI Spec for Istio APIs."
		}
		// derive the API version from the package name
		// which is a convention for Istio APIs.
		var p string
		if pkg != nil {
			p = pkg.GetPackage()
		} else {
			p = name
		}
		s := strings.Split(p, ".")
		version = s[len(s)-1]
	} else {
		description = "OpenAPI Spec for Istio APIs."
	}

	c := openapi3.NewComponents()
	c.Schemas = allSchemas
	// add the openapi object required by the spec.
	o := openapi3.Swagger{
		OpenAPI: "3.0.1",
		Info: openapi3.Info{
			Title:   description,
			Version: version,
		},
		Components: c,
	}

	g.buffer.Reset()
	var filename *string
	if g.yaml {
		b, err := yaml.Marshal(o)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to marshall the output of %v to yaml", name)
		}
		filename = proto.String(name + ".yaml")
		g.buffer.Write(b)
	} else {
		b, err := json.MarshalIndent(o, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to marshall the output of %v to json", name)
		}
		filename = proto.String(name + ".json")
		g.buffer.Write(b)
	}

	return plugin.CodeGeneratorResponse_File{
		Name:    filename,
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
			sr := g.fieldTypeRef(field)
			if g.useRef && sr.Ref != "" {
				// in `$ref`, the value of the schema is not in the output.
				sr.Value = nil
				o.WithPropertyRef(field.GetName(), sr)
			} else {
				o.WithProperty(field.GetName(), sr.Value)
			}
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
		if s, ok := specialTypes[g.absoluteName(msg)]; ok {
			schema = &s
		} else if msg.GetOptions().GetMapEntry() {
			isMap = true
			sr := g.fieldTypeRef(msg.Fields[1])
			if g.useRef && sr.Ref != "" {
				schema = openapi3.NewObjectSchema()
				// in `$ref`, the value of the schema is not in the output.
				sr.Value = nil
				schema.AdditionalProperties = sr
			} else {
				schema = openapi3.NewObjectSchema().WithAdditionalProperties(sr.Value)
			}
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

// fieldTypeRef generates the `$ref` in addition to the schema for a field.
func (g *openapiGenerator) fieldTypeRef(field *protomodel.FieldDescriptor) *openapi3.SchemaRef {
	s := g.fieldType(field)
	var ref string
	if *field.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		msg := field.FieldType.(*protomodel.MessageDescriptor)
		// only generate `$ref` for top level messages.
		if _, ok := g.messages[g.relativeName(field.FieldType)]; ok && msg.Parent == nil {
			ref = fmt.Sprintf("#/components/schemas/%v", g.absoluteName(field.FieldType))
		}
	}
	return openapi3.NewSchemaRef(ref, s)
}

func (g *openapiGenerator) relativeName(desc protomodel.CoreDesc) string {
	typeName := protomodel.DottedName(desc)
	if desc.PackageDesc() == g.currentPackage {
		return typeName
	}

	return desc.PackageDesc().Name + "." + typeName
}
